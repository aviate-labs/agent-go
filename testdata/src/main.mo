actor {
    var v = 0;

    public query func get() : async Nat { v };

    public shared func add(n : Nat) : async Nat {
        v += n;
        v;
    };
};
