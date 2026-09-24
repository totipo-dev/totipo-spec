import java.io.*;
import java.math.BigInteger;
import java.nio.*;
import java.nio.charset.*;
import java.nio.file.*;
import java.security.*;
import java.security.spec.*;
import java.util.*;
import javax.crypto.*;
import javax.crypto.spec.*;

/** r31 executable byte review. Test-only parser/encoder, not production code.
 * Two independent TLV framing readers share the registry validator. The field
 * oracle remains independent of wire framing and cryptographic processing.
 */
public class R31Review {
    static long checks; static int aeadCalls, parseCalls;
    static final HexFormat HEX=HexFormat.of();
    static final byte[] PUB=hex("d75a980182b10ab7d54bfed3c964073a0ee172f3daa62325af021a68f707511a");
    static final byte[] SEED=hex("9d61b19deffd5a60ba844af492ec2cc44449c5697b326919703bac031cae7f60");
    static final byte[] ROOT=new byte[32]; // Public deterministic fixture material only.
    static final byte[] PRK=mac(new byte[32],ROOT);
    static final byte[] KID=expand(PRK,ascii("TOTP-Vault/v0/object-id"));
    static final byte[] KOBJ=expand(PRK,ascii("TOTP-Vault/v0/object-key-root"));
    static final byte[] KSIG=expand(PRK,ascii("TOTP-Vault/v0/signature-context"));
    static class Bad extends RuntimeException { final String stage; Bad(String stage){super(stage);this.stage=stage;} }
    static void need(boolean b,String s){if(!b)throw new Bad(s);}
    static void check(boolean b,String s){checks++;if(!b)throw new AssertionError(s);}
    interface Action {void run() throws Exception;}
    static void bad(String stage,Action a)throws Exception{
        try{a.run();}catch(Bad e){check(e.stage.equals(stage),"expected "+stage+", got "+e.stage);return;}
        throw new AssertionError("accepted invalid "+stage);
    }
    static byte[] hex(String s){return HEX.parseHex(s);}
    static byte[] ascii(String s){return s.getBytes(StandardCharsets.US_ASCII);}
    static byte[] cat(byte[]... arrays){var b=new ByteArrayOutputStream();for(byte[] a:arrays)b.writeBytes(a);return b.toByteArray();}
    static byte[] num(long v,int width){byte[] a=new byte[width];for(int i=width-1;i>=0;i--){a[i]=(byte)v;v>>>=8;}return a;}
    static long number(byte[] b){long v=0;for(byte x:b)v=(v<<8)|(x&255);return v;}
    static byte[] mac(byte[] key,byte[] input){try{Mac m=Mac.getInstance("HmacSHA256");m.init(new SecretKeySpec(key,"HmacSHA256"));return m.doFinal(input);}catch(Exception e){throw new RuntimeException(e);}}
    static byte[] expand(byte[] key,byte[] info){return mac(key,cat(info,new byte[]{1}));}
    record Tlv(int tag,byte[] value){}
    static byte[] encode(List<Tlv> list){var b=new ByteArrayOutputStream();for(Tlv t:list)b.writeBytes(cat(num(t.tag,2),num(t.value.length,2),t.value));return b.toByteArray();}
    static Tlv t(int tag,byte[] value){return new Tlv(tag,value);}
    static List<Tlv> framing(byte[] bytes){
        List<Tlv> out=new ArrayList<>();ByteBuffer b=ByteBuffer.wrap(bytes);
        while(b.hasRemaining()){
            need(b.remaining()>=4,"TLV framing");int tag=Short.toUnsignedInt(b.getShort()),len=Short.toUnsignedInt(b.getShort());
            need(len<=b.remaining(),"TLV framing");byte[] v=new byte[len];b.get(v);out.add(t(tag,v));
        }return out;
    }
    static List<Tlv> framingStream(byte[] bytes)throws IOException{
        var in=new DataInputStream(new ByteArrayInputStream(bytes));List<Tlv> out=new ArrayList<>();
        while(in.available()!=0){
            need(in.available()>=4,"TLV framing");int tag=in.readUnsignedShort(),len=in.readUnsignedShort();
            need(len<=in.available(),"TLV framing");out.add(t(tag,in.readNBytes(len)));
        }return out;
    }
    static byte[] width(Tlv t,int n){need(t.value.length==n,"width");return t.value;}
    static String utf8(byte[] bytes){
        need(bytes.length<=256,"string limit");
        try{return StandardCharsets.UTF_8.newDecoder().onMalformedInput(CodingErrorAction.REPORT).onUnmappableCharacter(CodingErrorAction.REPORT).decode(ByteBuffer.wrap(bytes)).toString();}
        catch(CharacterCodingException e){throw new Bad("UTF-8");}
    }
    static void order(List<Tlv> list,boolean parents){
        int prev=-1;for(Tlv t:list){need(t.tag>prev||(parents&&t.tag==5&&prev==5),"tag order");prev=t.tag;}
    }
    static void credential(byte[] bytes){
        List<Tlv> ts=framing(bytes);order(ts,false);need(ts.size()==4,"credential tags");
        for(int i=0;i<4;i++)need(ts.get(i).tag==0x301+i,"credential tags");
        long algorithm=number(width(ts.get(0),1)),digits=number(width(ts.get(1),1)),period=number(width(ts.get(2),4));
        need(algorithm>=1&&algorithm<=3,"algorithm");need(digits>=6&&digits<=8,"digits");need(period>=1,"period");
        need(ts.get(3).value.length>=1&&ts.get(3).value.length<=128,"secret limit");
    }
    record Decoded(int type,List<String> parents,Map<String,String> fields,String token,byte[] pub,byte[] signature,byte[] unsigned){}
    static Decoded grammar(List<Tlv> list){
        order(list,true);Map<Integer,Tlv> tags=new HashMap<>();List<String> parents=new ArrayList<>();
        for(Tlv t:list){if(t.tag==5){width(t,32);parents.add(HEX.formatHex(t.value));}else tags.put(t.tag,t);}
        for(int tag:new int[]{1,2,3,4,0xff01})need(tags.containsKey(tag),"required tag");
        need(number(width(tags.get(1),1))==0,"version");int type=(int)number(width(tags.get(2),1));need(type==1||type==2,"object type");
        byte[] pub=width(tags.get(3),32),signature=width(tags.get(0xff01),64);
        long count=number(width(tags.get(4),2));need(count<=32,"parent limit");need(count==parents.size(),"parent count");
        for(int i=1;i<parents.size();i++)need(parents.get(i-1).compareTo(parents.get(i))<0,"parent order");
        Set<Integer> allowed=new HashSet<>(Set.of(1,2,3,4,5,0xff01));
        allowed.addAll(type==1?Set.of(0x101,0x102,0x103,0x104,0x105):Set.of(0x201));
        for(Tlv t:list)need(allowed.contains(t.tag),"unknown/forbidden tag");
        Map<String,String> fields=new HashMap<>();String token=null;
        if(type==1){
            need(tags.containsKey(0x101),"required tag");token=HEX.formatHex(width(tags.get(0x101),32));
            if(tags.containsKey(0x102)){long status=number(width(tags.get(0x102),1));need(status==1||status==2,"status");fields.put("STATUS",status==1?"LIVE":"TOMBSTONE");}
            for(int tag:new int[]{0x103,0x104})if(tags.containsKey(tag))fields.put(tag==0x103?"ISSUER":"ACCOUNT",utf8(tags.get(tag).value));
            if(tags.containsKey(0x105)){byte[] c=tags.get(0x105).value;credential(c);fields.put("CREDENTIAL",HEX.formatHex(c));}
            need(!fields.isEmpty(),"field required");
        }else{need(tags.containsKey(0x201),"required tag");fields.put("DISPLAY_NAME",utf8(tags.get(0x201).value));}
        return new Decoded(type,List.copyOf(parents),Map.copyOf(fields),token,pub,signature,encode(list.stream().filter(t->t.tag!=0xff01).toList()));
    }
    static Decoded decode(byte[] p,boolean stream)throws Exception{return grammar(stream?framingStream(p):framing(p));}
    static void both(byte[] p)throws Exception{Decoded a=decode(p,false),b=decode(p,true);check(a.type==b.type&&a.parents.equals(b.parents)&&a.fields.equals(b.fields)&&Objects.equals(a.token,b.token)&&Arrays.equals(a.unsigned,b.unsigned),"framing implementations disagree");}
    static void rejectBoth(byte[] p,String stage)throws Exception{bad(stage,()->decode(p,false));bad(stage,()->decode(p,true));}
    static List<Tlv> replace(List<Tlv> list,int tag,byte[] value){List<Tlv> out=new ArrayList<>();for(Tlv t:list)out.add(t.tag==tag?t(tag,value):t);return out;}
    static byte[] cred(int algorithm,int digits,long period,int secret){return encode(List.of(t(0x301,num(algorithm,1)),t(0x302,num(digits,1)),t(0x303,num(period,4)),t(0x304,new byte[secret])));}
    static List<Tlv> object(int type,int parents,boolean all){
        List<Tlv> ts=new ArrayList<>(List.of(t(1,num(0,1)),t(2,num(type,1)),t(3,PUB),t(4,num(parents,2))));
        for(int i=0;i<parents;i++)ts.add(t(5,num(i,32)));
        if(type==1){ts.add(t(0x101,new byte[32]));ts.add(t(0x102,num(1,1)));if(all){ts.add(t(0x103,ascii("issuer")));ts.add(t(0x104,ascii("account")));ts.add(t(0x105,cred(1,6,30,20)));}}
        else ts.add(t(0x201,ascii("device")));
        ts.add(t(0xff01,new byte[64]));return ts;
    }
    static byte[] signatureInput(int type,byte[] unsigned){return cat(ascii(type==1?"TOTP-Vault/v0/token-update":"TOTP-Vault/v0/device-update"),KSIG,unsigned);}
    static byte[] sign(List<Tlv> ts)throws Exception{
        byte[] unsigned=encode(ts.stream().filter(t->t.tag!=0xff01).toList());
        int type=(int)number(ts.stream().filter(t->t.tag==2).findFirst().orElseThrow().value);
        Signature s=Signature.getInstance("Ed25519");s.initSign(KeyFactory.getInstance("Ed25519").generatePrivate(new PKCS8EncodedKeySpec(cat(hex("302e020100300506032b657004220420"),SEED))));s.update(signatureInput(type,unsigned));return encode(replace(ts,0xff01,s.sign()));
    }
    static byte[] frame(byte[] p){need(p.length<=2030,"envelope length");return cat(num(p.length,2),p,new byte[2030-p.length]);}
    static byte[] unframe(byte[] plain){need(plain.length==2032,"plaintext length");int n=(int)number(Arrays.copyOf(plain,2));need(n<=2030,"envelope length");for(int i=2+n;i<plain.length;i++)need(plain[i]==0,"padding");return Arrays.copyOfRange(plain,2,2+n);}
    static byte[] objectKey(byte[] id){return expand(KOBJ,cat(ascii("TOTP-Vault/v0/object-key"),id));}
    static byte[] crypt(int mode,byte[] id,byte[] input)throws Exception{
        Cipher c=Cipher.getInstance("AES/GCM/NoPadding");c.init(mode,new SecretKeySpec(objectKey(id),"AES"),new GCMParameterSpec(128,Arrays.copyOf(id,12)));c.updateAAD(cat(ascii("TOTP-Vault/v0/object"),id));return c.doFinal(input);
    }
    record Stored(byte[] id,byte[] file){}
    static Stored store(byte[] p)throws Exception{byte[] id=mac(KID,p);return new Stored(id,crypt(Cipher.ENCRYPT_MODE,id,frame(p)));}
    record Read(boolean future,Decoded decoded){}
    static Read read(Stored s)throws Exception{
        need(s.file.length==2048,"file length");byte[] plain;
        aeadCalls++;try{plain=crypt(Cipher.DECRYPT_MODE,s.id,s.file);}catch(AEADBadTagException e){throw new Bad("AEAD");}
        byte[] p=unframe(plain);need(MessageDigest.isEqual(s.id,mac(KID,p)),"object ID");
        parseCalls++;
        // Decode only the stable prefix before version dispatch, leaving a future
        // version's remaining bytes opaque even if they are not valid v0 TLVs.
        need(p.length>=10,"prefix");need(number(Arrays.copyOfRange(p,0,2))==1&&number(Arrays.copyOfRange(p,2,4))==1
                &&number(Arrays.copyOfRange(p,5,7))==2&&number(Arrays.copyOfRange(p,7,9))==1,"prefix");
        if(p[4]!=0)return new Read(true,null);
        Decoded d=decode(p,false);
        need(R17Review.strict(d.pub,d.signature,signatureInput(d.type,d.unsigned)),"signature");return new Read(false,d);
    }
    static void registryTests()throws Exception{
        for(int type:new int[]{1,2})for(int count:new int[]{0,1,32})both(encode(object(type,count,true)));
        rejectBoth(encode(object(1,33,true)),"parent limit");
        var root=object(1,0,true);var parents=object(1,2,true);
        rejectBoth(encode(replace(parents,4,num(1,2))),"parent count");
        for(int n:new int[]{31,33})rejectBoth(encode(replace(parents,5,new byte[n])),"width");
        var duplicate=new ArrayList<>(parents);duplicate.set(5,duplicate.get(4));rejectBoth(encode(duplicate),"parent order");
        var reversed=new ArrayList<>(parents);Collections.swap(reversed,4,5);rejectBoth(encode(reversed),"parent order");
        for(int tag:new int[]{1,2,3,4,0x101,0x102,0xff01}){
            Tlv original=root.stream().filter(t->t.tag==tag).findFirst().orElseThrow();
            for(int n:new int[]{original.value.length-1,original.value.length+1})rejectBoth(encode(replace(root,tag,new byte[n])),"width");
        }
        var repeated=new ArrayList<>(root);repeated.add(1,repeated.get(0));rejectBoth(encode(repeated),"tag order");
        var unordered=new ArrayList<>(root);Collections.swap(unordered,0,1);rejectBoth(encode(unordered),"tag order");
        rejectBoth(cat(encode(root),new byte[]{0}),"TLV framing");
        byte[] truncated=encode(root);rejectBoth(Arrays.copyOf(truncated,truncated.length-1),"TLV framing");
        var absent=root.stream().filter(t->t.tag!=3).toList();rejectBoth(encode(absent),"required tag");
        rejectBoth(encode(root.stream().filter(t->t.tag!=0xff01).toList()),"required tag");
        rejectBoth(encode(root.stream().filter(t->t.tag<0x102||t.tag>0x105).toList()),"field required");
        var wrongDevice=new ArrayList<>(object(2,0,true));wrongDevice.add(t(0x101,new byte[32]));
        wrongDevice.sort(Comparator.comparingInt(Tlv::tag));rejectBoth(encode(wrongDevice),"unknown/forbidden tag");
        rejectBoth(encode(replace(root,2,num(3,1))),"object type");
        rejectBoth(encode(replace(root,0x102,num(3,1))),"status");
        for(int tag:new int[]{6,0x201,0xff02}){
            var unknown=new ArrayList<>(root);unknown.add(t(tag,new byte[0]));unknown.sort(Comparator.comparingInt(Tlv::tag));rejectBoth(encode(unknown),"unknown/forbidden tag");
        }
        var moved=new ArrayList<>(root);Tlv sig=moved.remove(moved.size()-1);moved.add(4,sig);rejectBoth(encode(moved),"tag order");
        for(int tag:new int[]{0x103,0x104,0x201}){
            var base=object(tag==0x201?2:1,0,true);
            for(int n:new int[]{0,256})both(encode(replace(base,tag,new byte[n])));
            byte[] multibyte="é".repeat(128).getBytes(StandardCharsets.UTF_8);both(encode(replace(base,tag,multibyte)));
            rejectBoth(encode(replace(base,tag,new byte[257])),"string limit");
            for(byte[] broken:List.of(hex("c080"),hex("eda080"),hex("f4908080"),hex("80")))rejectBoth(encode(replace(base,tag,broken)),"UTF-8");
        }
        for(int algorithm=1;algorithm<=3;algorithm++)for(int digits=6;digits<=8;digits++)for(long period:new long[]{1,30,0x80000000L,0xffffffffL})
            for(int length:new int[]{1,128})both(encode(replace(root,0x105,cred(algorithm,digits,period,length))));
        for(int length:new int[]{0,129})rejectBoth(encode(replace(root,0x105,cred(1,6,30,length))),"secret limit");
        for(int algorithm:new int[]{0,4,255})rejectBoth(encode(replace(root,0x105,cred(algorithm,6,30,20))),"algorithm");
        for(int digits:new int[]{0,5,9,255})rejectBoth(encode(replace(root,0x105,cred(1,digits,30,20))),"digits");
        rejectBoth(encode(replace(root,0x105,cred(1,6,0,20))),"period");
        var nested=framing(cred(1,6,30,20));
        for(int tag:new int[]{0x301,0x302,0x303}){
            int width=tag==0x303?4:1;
            for(int n:new int[]{width-1,width+1})rejectBoth(encode(replace(root,0x105,encode(replace(nested,tag,new byte[n])))),"width");
        }
        var nestedOrder=new ArrayList<>(nested);Collections.swap(nestedOrder,0,1);rejectBoth(encode(replace(root,0x105,encode(nestedOrder))),"tag order");
        var nestedDuplicate=new ArrayList<>(nested);nestedDuplicate.add(1,nestedDuplicate.get(0));rejectBoth(encode(replace(root,0x105,encode(nestedDuplicate))),"tag order");
        var nestedUnknown=new ArrayList<>(nested);nestedUnknown.add(t(0x305,new byte[0]));rejectBoth(encode(replace(root,0x105,encode(nestedUnknown))),"credential tags");
        var replacedNested=new ArrayList<>(nested);replacedNested.set(3,t(0x305,new byte[20]));rejectBoth(encode(replace(root,0x105,encode(replacedNested))),"credential tags");
        rejectBoth(encode(replace(root,0x105,cat(encode(nested),new byte[]{0}))),"TLV framing");
        for(int type:new int[]{1,2}){
            var max=object(type,32,true);
            if(type==1){max=replace(max,0x103,new byte[256]);max=replace(max,0x104,new byte[256]);max=replace(max,0x105,cred(3,8,0xffffffffL,128));}
            else max=replace(max,0x201,new byte[256]);
            byte[] p=sign(max);check(p.length==(type==1?1987:1532),"maximum semantic size");both(p);check(!read(store(p)).future,"max object authentication");
        }
    }
    static void envelopeTests()throws Exception{
        byte[] p=sign(object(1,0,true));Stored s=store(p);check(Arrays.equals(s.file,store(p).file),"deterministic encryption");
        check(Arrays.equals(read(s).decoded.unsigned,decode(p,false).unsigned),"authenticated round trip");
        for(int n:new int[]{0,1}){
            byte[] small=new byte[n];check(Arrays.equals(unframe(frame(small)),small),"small framing success");bad("prefix",()->read(store(small)));
        }
        byte[] generic=new byte[2030];check(unframe(frame(generic)).length==2030,"generic capacity");bad("prefix",()->read(store(generic)));
        byte[] future=cat(encode(List.of(t(1,num(1,1)),t(2,num(255,1)))),new byte[2020]);
        check(read(store(future)).future,"future maximum bypasses v0 size/type/grammar rules");
        byte[] futureBrokenTail=cat(Arrays.copyOf(future,10),new byte[]{(byte)255});check(read(store(futureBrokenTail)).future,"opaque future tail");
        byte[] malformedPrefix=future.clone();malformedPrefix[3]=2;bad("prefix",()->read(store(malformedPrefix)));
        bad("object type",()->read(store(encode(replace(object(1,0,true),2,num(255,1))))));
        bad("envelope length",()->frame(new byte[2031]));
        for(int n:new int[]{2031,65535}){
            byte[] wrong=frame(p);System.arraycopy(num(n,2),0,wrong,0,2);Stored corrupt=new Stored(s.id,crypt(Cipher.ENCRYPT_MODE,s.id,wrong));
            int calls=parseCalls;bad("envelope length",()->read(corrupt));check(parseCalls==calls,"oversized length reached parser");
        }
        byte[] padding=frame(p);padding[padding.length-1]=1;Stored padded=new Stored(s.id,crypt(Cipher.ENCRYPT_MODE,s.id,padding));
        int calls=parseCalls;bad("padding",()->read(padded));check(parseCalls==calls,"padding reached parser");
        for(int n:new int[]{0,2047,2049,4096}){
            int aead=aeadCalls;bad("file length",()->read(new Stored(s.id,Arrays.copyOf(s.file,n))));check(aeadCalls==aead,"wrong file size reached AEAD");
        }
        byte[] tampered=s.file.clone();tampered[15]^=1;calls=parseCalls;bad("AEAD",()->read(new Stored(s.id,tampered)));check(parseCalls==calls,"unauthenticated plaintext parsed");
        byte[] otherId=s.id.clone();otherId[31]^=1;bad("AEAD",()->read(new Stored(otherId,s.file)));
        Stored wrongIdentity=new Stored(otherId,crypt(Cipher.ENCRYPT_MODE,otherId,frame(p)));
        calls=parseCalls;bad("object ID",()->read(wrongIdentity));check(parseCalls==calls,"ID mismatch reached parser");
        byte[] invalidSig=p.clone();invalidSig[invalidSig.length-1]^=1;bad("signature",()->read(store(invalidSig)));
        byte[] unsigned=decode(p,false).unsigned;check(unsigned.length==p.length-68,"unsigned form omits entire terminal TLV");
        check(!R17Review.strict(PUB,decode(p,false).signature,signatureInput(2,unsigned)),"signature type domain separation");
        byte[] otherVault=signatureInput(1,unsigned);otherVault[ascii("TOTP-Vault/v0/token-update").length]^=1;
        check(!R17Review.strict(PUB,decode(p,false).signature,otherVault),"signature vault binding");
    }
    static String otp(String algo,byte[] key,BigInteger seconds,long period,int digits)throws Exception{
        need(period>=1&&period<=0xffffffffL,"period");need(seconds.signum()>=0,"time");need(digits>=6&&digits<=8,"digits");
        BigInteger counter=seconds.divide(BigInteger.valueOf(period));need(counter.bitLength()<=64,"counter");
        byte[] moving=num(counter.longValue(),8);Mac m=Mac.getInstance(algo);m.init(new SecretKeySpec(key,algo));byte[] h=m.doFinal(moving);
        int offset=h[h.length-1]&15;long truncated=ByteBuffer.wrap(h,offset,4).getInt()&0x7fffffffL;
        return String.format(Locale.ROOT,"%0"+digits+"d",truncated%(long)Math.pow(10,digits));
    }
    static void periods()throws Exception{
        byte[] key=ascii("12345678901234567890");
        // RFC 4226 counter 0/1/2 values anchor period boundaries independently of
        // division in the implementation under test.
        String[] hotp={"755224","287082","359152"};
        for(long period:new long[]{1,30,0x80000000L,0xffffffffL}){
            for(int c=0;c<3;c++){
                BigInteger start=BigInteger.valueOf(period).multiply(BigInteger.valueOf(c));
                check(otp("HmacSHA1",key,start,period,6).equals(hotp[c]),"period start");
                check(otp("HmacSHA1",key,start.add(BigInteger.valueOf(period-1)),period,6).equals(hotp[c]),"period end");
                check(R17Review.totp("HmacSHA1",key,start.longValue(),period,6).equals(hotp[c]),"legacy helper unsigned period range");
            }
        }
        for(String algorithm:List.of("HmacSHA1","HmacSHA256","HmacSHA512"))for(int digits=6;digits<=8;digits++){
            String expected=R17Review.totp(algorithm,key,59,30,digits);
            check(otp(algorithm,key,BigInteger.valueOf(0xffffffffL),0xffffffffL,digits).equals(expected),"large period counter one");
        }
        bad("period",()->otp("HmacSHA1",key,BigInteger.ZERO,0,6));bad("period",()->otp("HmacSHA1",key,BigInteger.ZERO,0x100000000L,6));
        bad("time",()->otp("HmacSHA1",key,BigInteger.valueOf(-1),30,6));
        BigInteger max=BigInteger.ONE.shiftLeft(64).subtract(BigInteger.ONE);
        check(otp("HmacSHA1",key,max,1,6).length()==6,"unsigned counter maximum");
        bad("counter",()->otp("HmacSHA1",key,max.add(BigInteger.ONE),1,6));
        byte[] high=cred(1,6,0xffffffffL,20);credential(high);check(number(framing(high).get(2).value)==0xffffffffL,"warning policy cannot reject in-range period");
    }
    static List<Tlv> tokenFields(List<byte[]> parents,Map<Integer,byte[]> values){
        List<Tlv> list=new ArrayList<>(List.of(t(1,num(0,1)),t(2,num(1,1)),t(3,PUB),t(4,num(parents.size(),2))));
        parents.stream().sorted(Arrays::compareUnsigned).forEach(p->list.add(t(5,p)));
        list.add(t(0x101,new byte[32]));values.entrySet().stream().sorted(Map.Entry.comparingByKey()).forEach(e->list.add(t(e.getKey(),e.getValue())));
        list.add(t(0xff01,new byte[64]));return list;
    }
    static void bridge()throws Exception{
        List<R26Oracle.Event> raw=new ArrayList<>();Map<String,Integer> identities=new HashMap<>();Map<String,Integer> tokens=new HashMap<>();
        Stored root=store(sign(object(1,0,true)));
        Stored del=store(sign(tokenFields(List.of(root.id),Map.of(0x102,num(2,1)))));
        Stored rotation=store(sign(tokenFields(List.of(root.id),Map.of(0x105,cred(2,8,0xffffffffL,32)))));
        Stored resolved=store(sign(tokenFields(List.of(del.id,rotation.id),Map.of(0x102,num(1,1)))));
        for(Stored s:List.of(root,del,rotation,resolved)){
            Decoded d=read(s).decoded;Set<Integer> cp=new HashSet<>();for(String id:d.parents){need(identities.containsKey(id),"bridge pending");cp.add(identities.get(id));}
            int id=identities.size();identities.put(HEX.formatHex(s.id),id);int token=tokens.computeIfAbsent(d.token,x->tokens.size());
            raw.add(new R26Oracle.Event(id,token,cp,d.fields));
        }
        R26Oracle oracle=new R26Oracle(raw);
        check(oracle.view(Set.of(1,2)).conflicts().equals(Set.of(new R26Oracle.Pair(1,2))),"decoded race differs from symbolic oracle");
        check(oracle.view(Set.of(3)).conflicts().isEmpty(),"decoded resolution differs from symbolic oracle");
        // A capacity-blocked writer must not confuse a graph-valid >32 frontier
        // with an encodable one. Signatures/framing for these artificial identities
        // are outside this separate capacity fixture.
        R26Review.Model model=new R26Review.Model();int r=model.root(0);Set<Integer> heads=new HashSet<>();List<byte[]> parentIds=new ArrayList<>();
        for(int i=1;i<=33;i++){
            heads.add(model.add(0,Set.of(r),Map.of("ACCOUNT","a"+i)));parentIds.add(num(i,32));
            if(i==32){
                both(encode(tokenFields(parentIds,Map.of(0x104,ascii("merged")))));
                check(new R26Oracle(model.raw()).view(heads).fields().get("ACCOUNT").size()==32,"32-head frontier");
            }
        }
        check(new R26Oracle(model.raw()).view(heads).fields().get("ACCOUNT").size()==33,"graph preserves 33 alternatives");
        rejectBoth(encode(tokenFields(parentIds,Map.of(0x104,ascii("merged")))),"parent limit");
        check(heads.size()==33,"capacity failure must not truncate frontier");
    }
    static String vectors()throws Exception{
        StringBuilder out=new StringBuilder("# Public deterministic r31 review fixtures; NOT real vault secrets.\n");
        for(int type:new int[]{1,2}){
            byte[] p=sign(object(type,0,true));Decoded d=decode(p,false);Stored s=store(p);check(!read(s).future,"fixture decrypt/validate");
            out.append("\n[").append(type==1?"TOKEN_UPDATE":"DEVICE_UPDATE").append("]\n");
            Map<String,byte[]> fields=new LinkedHashMap<>();fields.put("root",ROOT);fields.put("signing_seed",SEED);fields.put("public_key",PUB);
            fields.put("prk",PRK);fields.put("k_id",KID);fields.put("k_object_root",KOBJ);fields.put("k_signature_context",KSIG);
            fields.put("unsigned",d.unsigned);fields.put("signature_input",signatureInput(type,d.unsigned));fields.put("signature",d.signature);
            fields.put("signed_plaintext",p);fields.put("object_id",s.id);fields.put("object_key",objectKey(s.id));fields.put("nonce",Arrays.copyOf(s.id,12));
            fields.put("aad",cat(ascii("TOTP-Vault/v0/object"),s.id));fields.put("envelope_plaintext",frame(p));fields.put("ciphertext",Arrays.copyOf(s.file,2032));fields.put("tag",Arrays.copyOfRange(s.file,2032,2048));
            fields.forEach((k,v)->out.append(k).append('=').append(HEX.formatHex(v)).append('\n'));
        }return out.toString();
    }
    public static void main(String[] args)throws Exception{
        Path fixture=Path.of("r31-object-vectors.txt");String actual=vectors();
        if(args.length==1&&args[0].equals("--write-vectors")){Files.writeString(fixture,actual);System.out.println("Wrote public deterministic object vectors.");return;}
        check(Files.readString(fixture).equals(actual),"pinned object-crypto vectors changed");
        registryTests();envelopeTests();periods();bridge();
        System.out.printf(Locale.ROOT,"R31 PASS: %,d checks; pinned TOKEN_UPDATE/DEVICE_UPDATE crypto vectors; two TLV framing readers; registry/envelope negatives; unsigned-period boundaries; decoded lifecycle bridge.%n",checks);
        System.out.println("Scope: shared registry validator, independent framing readers, existing symbolic lifecycle oracle. Not two independent full wire validators, bootstrap parser, real persistence/UI tests, or complete conformance corpus.");
    }
}
