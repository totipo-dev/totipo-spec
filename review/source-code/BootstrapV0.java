import java.nio.*;
import java.nio.charset.*;
import java.security.*;
import java.util.*;
import javax.crypto.*;
import javax.crypto.spec.*;
import org.bouncycastle.crypto.generators.Argon2BytesGenerator;
import org.bouncycastle.crypto.params.Argon2Parameters;

/** r32 in-memory bootstrap codec. Does not install files, pin vault identity, or
 * implement the durable establishment protocol. Salt/nonce freshness is the
 * responsibility of the caller using the explicit deterministic wrap method.
 */
public final class BootstrapV0 {
    public static final int LENGTH=87, HEADER_LENGTH=39, PASSWORD_LIMIT=1024;
    private static final byte[] MAGIC="TOTP-VAULT".getBytes(StandardCharsets.US_ASCII);
    @FunctionalInterface public interface Kdf { byte[] derive(byte[] password,byte[] salt); }
    private final Kdf kdf;
    public BootstrapV0(){this(BootstrapV0::argon2id);}
    // Injection permits counting KDF invocations in pre-KDF rejection tests.
    BootstrapV0(Kdf kdf){this.kdf=Objects.requireNonNull(kdf);}

    static byte[] argon2id(byte[] password,byte[] salt){
        if(password.length>PASSWORD_LIMIT||salt.length!=16)throw new IllegalArgumentException("KDF input length");
        var builder=new Argon2Parameters.Builder(Argon2Parameters.ARGON2_id)
                .withVersion(Argon2Parameters.ARGON2_VERSION_13).withMemoryAsKB(65536)
                .withIterations(3).withParallelism(4).withSalt(salt)
                .withSecret(new byte[0]).withAdditional(new byte[0]);
        var parameters=builder.build();
        try{
            var generator=new Argon2BytesGenerator();generator.init(parameters);
            byte[] out=new byte[32];generator.generateBytes(password,out);return out;
        }finally{parameters.clear();builder.clear();}
    }
    public static byte[] passwordBytes(char[] password){
        if(password.length>PASSWORD_LIMIT)throw new IllegalArgumentException("password exceeds UTF-8 byte limit");
        ByteBuffer encoded=null;
        try{
            encoded=StandardCharsets.UTF_8.newEncoder().onMalformedInput(CodingErrorAction.REPORT)
                    .onUnmappableCharacter(CodingErrorAction.REPORT).encode(CharBuffer.wrap(password));
            if(encoded.remaining()>PASSWORD_LIMIT)throw new IllegalArgumentException("password exceeds UTF-8 byte limit");
            byte[] out=new byte[encoded.remaining()];encoded.get(out);return out;
        }catch(CharacterCodingException e){throw new IllegalArgumentException("invalid Unicode password",e);}
        finally{if(encoded!=null&&encoded.hasArray())Arrays.fill(encoded.array(),(byte)0);}
    }
    private static void passwordCheck(byte[] password){
        if(password.length>PASSWORD_LIMIT)throw new IllegalArgumentException("password exceeds UTF-8 byte limit");
        CharBuffer decoded=null;
        try{decoded=StandardCharsets.UTF_8.newDecoder().onMalformedInput(CodingErrorAction.REPORT)
                    .onUnmappableCharacter(CodingErrorAction.REPORT).decode(ByteBuffer.wrap(password));}
        catch(CharacterCodingException e){throw new IllegalArgumentException("password is not UTF-8",e);}
        finally{if(decoded!=null&&decoded.hasArray())Arrays.fill(decoded.array(),'\0');}
    }
    private static void framing(byte[] file){
        if(file.length!=LENGTH)throw new IllegalArgumentException("bootstrap length must be 87");
        for(int i=0;i<MAGIC.length;i++)if(file[i]!=MAGIC[i])throw new IllegalArgumentException("bootstrap magic");
        if(file[10]!=0)throw new IllegalArgumentException("unsupported bootstrap version");
    }
    public byte[] unlock(byte[] file,char[] password)throws GeneralSecurityException{
        // Framing must be checked even when the caller supplied malformed text.
        framing(file);byte[] bytes=passwordBytes(password);
        try{return unlockBytes(file,bytes);}finally{Arrays.fill(bytes,(byte)0);}
    }
    public byte[] unlockBytes(byte[] file,byte[] password)throws GeneralSecurityException{
        if(file.length!=LENGTH)throw new IllegalArgumentException("bootstrap length must be 87");
        byte[] snapshot=file.clone();framing(snapshot);passwordCheck(password);
        byte[] pw=password.clone(),key=null;
        try{
            key=kdf.derive(pw,Arrays.copyOfRange(snapshot,11,27));
            if(key.length!=32)throw new GeneralSecurityException("KDF output length");
            Cipher cipher=Cipher.getInstance("AES/GCM/NoPadding");
            cipher.init(Cipher.DECRYPT_MODE,new SecretKeySpec(key,"AES"),new GCMParameterSpec(128,snapshot,27,12));
            cipher.updateAAD(snapshot,0,HEADER_LENGTH);
            // doFinal authenticates before any root bytes are returned.
            return cipher.doFinal(snapshot,HEADER_LENGTH,LENGTH-HEADER_LENGTH);
        }finally{Arrays.fill(pw,(byte)0);if(key!=null)Arrays.fill(key,(byte)0);}
    }
    public byte[] wrap(byte[] root,char[] password,SecureRandom random)throws GeneralSecurityException{
        byte[] salt=new byte[16],nonce=new byte[12];random.nextBytes(salt);random.nextBytes(nonce);
        byte[] bytes=passwordBytes(password);
        try{return wrapBytes(root,bytes,salt,nonce);}finally{Arrays.fill(bytes,(byte)0);}
    }
    /** Deterministic entry point for vector consumption. Production callers use
     * wrap(root,password,random) to generate a fresh salt and nonce each time. */
    public byte[] wrapBytes(byte[] root,byte[] password,byte[] salt,byte[] nonce)throws GeneralSecurityException{
        if(root.length!=32||salt.length!=16||nonce.length!=12)throw new IllegalArgumentException("root/salt/nonce length");
        passwordCheck(password);byte[] header=new byte[HEADER_LENGTH];
        System.arraycopy(MAGIC,0,header,0,10);System.arraycopy(salt,0,header,11,16);System.arraycopy(nonce,0,header,27,12);
        byte[] pw=password.clone(),key=null;
        try{
            key=kdf.derive(pw,Arrays.copyOfRange(header,11,27));
            if(key.length!=32)throw new GeneralSecurityException("KDF output length");
            Cipher cipher=Cipher.getInstance("AES/GCM/NoPadding");
            cipher.init(Cipher.ENCRYPT_MODE,new SecretKeySpec(key,"AES"),new GCMParameterSpec(128,header,27,12));cipher.updateAAD(header);
            byte[] result=Arrays.copyOf(header,LENGTH);byte[] wrapped=cipher.doFinal(root);System.arraycopy(wrapped,0,result,HEADER_LENGTH,48);return result;
        }finally{Arrays.fill(pw,(byte)0);if(key!=null)Arrays.fill(key,(byte)0);}
    }
}
