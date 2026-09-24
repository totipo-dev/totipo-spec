import java.util.*;

/** Independent graph calculation for the r20 review, over already-valid histories.
 * Uses raw edges and Boolean transitive-closure matrices. It must not call the
 * incremental model, read its ancestry caches, or reuse its state/JOIN helpers.
 * This is a semantic oracle, not a wire parser or a full update validator.
 */
final class LifecycleOracle {
    record Event(int id, int token, Set<Integer> causal, Set<Integer> metadata,
                 Set<Integer> secret, Boolean live, boolean hasSecret, Integer confirmation) {
        Event {
            causal = Set.copyOf(causal);
            metadata = Set.copyOf(metadata);
            secret = Set.copyOf(secret);
        }
    }
    record Pair(int metadata, int witness) {}
    record View(Set<Integer> metadata, Set<Integer> secret, Set<Integer> frontier, Set<Pair> conflicts) {}
    enum Mutation { NONE, HEAD_WITNESSES_ONLY, CONFIRM_DESCENDANTS, NO_METADATA_PROPAGATION }

    private final List<Event> events;
    private final Map<Integer, Integer> index = new HashMap<>();
    private final boolean[][] causal, metadata, secret;
    private final boolean[] transitions;

    LifecycleOracle(Collection<Event> input) {
        events = input.stream().sorted(Comparator.comparingInt(Event::id)).toList();
        int n = events.size();
        causal = new boolean[n][n]; metadata = new boolean[n][n]; secret = new boolean[n][n];
        transitions = new boolean[n];
        for (int i = 0; i < n; i++) {
            if (index.put(events.get(i).id, i) != null) throw new IllegalArgumentException("duplicate identity");
        }
        for (int child = 0; child < n; child++) {
            Event e = events.get(child);
            edges(causal, e.causal, child);
            edges(metadata, e.metadata, child);
            edges(secret, e.secret, child);
            for (int parent : e.metadata) {
                Event p = events.get(position(parent));
                if (e.live == null || p.live == null) throw new IllegalArgumentException("untyped metadata edge");
                if (!e.live.equals(p.live)) transitions[child] = true;
            }
            for (int parent : e.secret)
                if (!e.hasSecret || !events.get(position(parent)).hasSecret)
                    throw new IllegalArgumentException("untyped secret edge");
            if (e.confirmation != null) {
                Event witness = events.get(position(e.confirmation));
                if (!witness.hasSecret || witness.token != e.token)
                    throw new IllegalArgumentException("invalid witness type/token");
            }
        }
        close(causal); close(metadata); close(secret);
        for (int i = 0; i < n; i++) for (int j = 0; j < n; j++) {
            if (i != j && causal[i][j] && causal[j][i]) throw new IllegalArgumentException("cycle");
            if ((metadata[i][j] || secret[i][j]) && !causal[i][j])
                throw new IllegalArgumentException("typed edge outside causal context");
        }
    }
    private int position(int id) {
        Integer pos = index.get(id);
        if (pos == null) throw new IllegalArgumentException("missing object " + id);
        return pos;
    }
    private void edges(boolean[][] relation, Set<Integer> parents, int child) {
        relation[child][child] = true;
        for (int id : parents) {
            int parent = position(id);
            if (events.get(parent).token != events.get(child).token)
                throw new IllegalArgumentException("cross-token edge");
            relation[parent][child] = true;
        }
    }
    private static void close(boolean[][] relation) {
        for (int k = 0; k < relation.length; k++)
            for (int i = 0; i < relation.length; i++) if (relation[i][k])
                for (int j = 0; j < relation.length; j++) relation[i][j] |= relation[k][j];
    }
    private Set<Integer> maxima(boolean[] visible, boolean[][] relation) {
        Set<Integer> result = new TreeSet<>();
        for (int i = 0; i < events.size(); i++) if (visible[i]) {
            boolean dominated = false;
            for (int j = 0; j < events.size(); j++)
                if (i != j && visible[j] && relation[i][j]) dominated = true;
            if (!dominated) result.add(events.get(i).id);
        }
        return result;
    }
    Set<Integer> fullyAvailable(Set<Integer> available) {
        Set<Integer> result = new TreeSet<>();
        for (int id : available) {
            int i = position(id); boolean complete = true;
            for (int j = 0; j < events.size(); j++)
                if (causal[j][i] && !available.contains(events.get(j).id)) complete = false;
            if (complete) result.add(id);
        }
        return result;
    }
    View view(Collection<Integer> roots) { return view(roots, Mutation.NONE); }
    View view(Collection<Integer> roots, Mutation mutation) {
        int n = events.size();
        boolean[] visible = new boolean[n], mh = new boolean[n], sh = new boolean[n];
        Integer token = null;
        for (int id : roots) {
            int root = position(id);
            if (token != null && token != events.get(root).token)
                throw new IllegalArgumentException("mixed-token query");
            token = events.get(root).token;
            for (int i = 0; i < n; i++) visible[i] |= causal[i][root];
        }
        for (int i = 0; i < n; i++) {
            mh[i] = visible[i] && events.get(i).live != null;
            sh[i] = visible[i] && events.get(i).hasSecret;
        }
        Set<Integer> metadataHeads = maxima(mh, metadata), secretHeads = maxima(sh, secret);
        Set<Pair> conflicts = new HashSet<>();
        // Accumulate all transition/witness races, then subtract confirmation
        // coverage, then take maximal witnesses. No calls to a cached ancestor set.
        for (int headId : metadataHeads) {
            int head = position(headId); boolean[] unresolved = new boolean[n];
            for (int l = 0; l < n; l++) if (metadata[l][head] && transitions[l]) {
                for (int w = 0; w < n; w++) if (sh[w]) {
                    if (mutation == Mutation.HEAD_WITNESSES_ONLY && !secretHeads.contains(events.get(w).id)) continue;
                    if (!causal[l][w] && !causal[w][l]) unresolved[w] = true;
                }
            }
            for (int c = 0; c < n; c++) if (metadata[c][head] && events.get(c).confirmation != null) {
                if (mutation == Mutation.NO_METADATA_PROPAGATION && c != head) continue;
                int q = position(events.get(c).confirmation);
                for (int w = 0; w < n; w++) {
                    boolean covered = secret[w][q];
                    if (mutation == Mutation.CONFIRM_DESCENDANTS) covered |= secret[q][w];
                    if (covered) unresolved[w] = false;
                }
            }
            for (int w : maxima(unresolved, secret)) conflicts.add(new Pair(headId, w));
        }
        return new View(Set.copyOf(metadataHeads), Set.copyOf(secretHeads),
                Set.copyOf(maxima(visible, causal)), Set.copyOf(conflicts));
    }
}
