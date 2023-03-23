actor {
  public type Tokens = {
    e8s : Nat64;
  };

  public type AccountIdentifier = Blob;

  public type AccountBalanceArgs = {
    account : AccountIdentifier;
  };

  public query func account_balance(args : AccountBalanceArgs) : async Tokens {
    { e8s = 1 };
  };

  type TimeStamp = {
    timestamp_nanos: Nat64;
};

  type Memo = Nat64;

  type SubAccount = Blob;

  type BlockIndex = Nat64;

  type TransferArgs = {
    memo : Memo;
    amount : Tokens;
    fee : Tokens;
    from_subaccount : ?SubAccount;
    to : AccountIdentifier;
    created_at_time : ?TimeStamp;
  };

  type TransferError = {
    #BadFee : { expected_fee : Tokens; };
    #InsufficientFunds : { balance: Tokens; };
    #TxTooOld : { allowed_window_nanos: Nat64 };
    #TxCreatedInFuture;
    #TxDuplicate : { duplicate_of: BlockIndex; }
};

  type TransferResult = {
    #Ok : BlockIndex;
    #Err : TransferError;
};

  public shared func transfer(args : TransferArgs) : async TransferResult { #Ok(1) };
};
