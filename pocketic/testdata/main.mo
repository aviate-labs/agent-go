actor {
    public query func helloQuery(name : Text) : async Text {
        "Hello, " # name # "!"
    };

    public shared func helloUpdate(name : Text) : async Text {
        "Hello, " # name # "!"
    };
};
