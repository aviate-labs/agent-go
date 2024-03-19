// Do NOT edit this file. It was automatically generated by https://github.com/aviate-labs/agent-go.
module T {
    public type EventKind = { #CyclesSent : { to : Principal; amount : Nat64; refund : Nat64 }; #CyclesReceived : { from : Principal; amount : Nat64; memo : ?Text }; #AddressAdded : { id : Principal; name : ?Text; role : T.Role }; #AddressRemoved : { id : Principal }; #CanisterCreated : { canister : Principal; cycles : Nat64 }; #CanisterCalled : { canister : Principal; method_name : Text; cycles : Nat64 }; #WalletDeployed : { canister : Principal } };
    public type EventKind128 = { #CyclesSent : { to : Principal; amount : Nat; refund : Nat }; #CyclesReceived : { from : Principal; amount : Nat; memo : ?Text }; #AddressAdded : { id : Principal; name : ?Text; role : T.Role }; #AddressRemoved : { id : Principal }; #CanisterCreated : { canister : Principal; cycles : Nat }; #CanisterCalled : { canister : Principal; method_name : Text; cycles : Nat }; #WalletDeployed : { canister : Principal } };
    public type Event = { id : Nat32; timestamp : Nat64; kind : T.EventKind };
    public type Event128 = { id : Nat32; timestamp : Nat64; kind : T.EventKind128 };
    public type Role = { #Contact; #Custodian; #Controller };
    public type Kind = { #Unknown; #User; #Canister };
    public type AddressEntry = { id : Principal; name : ?Text; kind : T.Kind; role : T.Role };
    public type ManagedCanisterInfo = { id : Principal; name : ?Text; created_at : Nat64 };
    public type ManagedCanisterEventKind = { #CyclesSent : { amount : Nat64; refund : Nat64 }; #Called : { method_name : Text; cycles : Nat64 }; #Created : { cycles : Nat64 } };
    public type ManagedCanisterEventKind128 = { #CyclesSent : { amount : Nat; refund : Nat }; #Called : { method_name : Text; cycles : Nat }; #Created : { cycles : Nat } };
    public type ManagedCanisterEvent = { id : Nat32; timestamp : Nat64; kind : T.ManagedCanisterEventKind };
    public type ManagedCanisterEvent128 = { id : Nat32; timestamp : Nat64; kind : T.ManagedCanisterEventKind128 };
    public type ReceiveOptions = { memo : ?Text };
    public type WalletResultCreate = { #Ok : { canister_id : Principal }; #Err : Text };
    public type WalletResult = { #Ok : (); #Err : Text };
    public type WalletResultCall = { #Ok : { return : Blob }; #Err : Text };
    public type CanisterSettings = { controller : ?Principal; controllers : ?[Principal]; compute_allocation : ?Nat; memory_allocation : ?Nat; freezing_threshold : ?Nat };
    public type CreateCanisterArgs = { cycles : Nat64; settings : T.CanisterSettings };
    public type CreateCanisterArgs128 = { cycles : Nat; settings : T.CanisterSettings };
    public type HeaderField = (Text, Text);
    public type HttpRequest = { method : Text; url : Text; headers : [T.HeaderField]; body : Blob };
    public type HttpResponse = { status_code : Nat16; headers : [T.HeaderField]; body : Blob; streaming_strategy : ?T.StreamingStrategy };
    public type StreamingCallbackHttpResponse = { body : Blob; token : ?T.Token };
    public type Token = {  };
    public type StreamingStrategy = { #Callback : { callback : { /* func */ }; token : T.Token } };
};
