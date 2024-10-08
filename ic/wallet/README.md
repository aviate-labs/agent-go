##  2 methods to access a canister:
- agent: directly access, normal purpose.
- wallet: use wallet canister as proxy, user can provide cycles for canister execution, demonstrating in Test_WalletCanister_LocalNet()

## Prerequisite
For test agent-go/agent_test.go/Test_Agent_LocalNet() and agent-go/ic/wallet/agent_test.go/Test_WalletCanister_LocalNet(), a hasher canister deployed on local replica for demonstration, its code shown below.
```rust

use ic_cdk_macros::{init, post_upgrade, pre_upgrade, query, update};
use ic_cdk::api::call::{msg_cycles_accept, msg_cycles_available};
use sha2::{Digest, Sha256};

const SINGLE_SHA256_COST_CYCLES :u64 = 200_000_000; 
const CONCAT_STRING_COST_CYCLES :u64 = 200_000_000; 

#[query]
pub fn greet()->String{
    format!("Hello, World!")
}

#[update]
pub fn sha256_with_cycles(msg :String, n :u32) -> String {
    let available_cycles = msg_cycles_available();
    let needed_cycles:u64 = SINGLE_SHA256_COST_CYCLES*(n as u64);
    ic_cdk::println!("needed_cycles: {}", needed_cycles);
    if available_cycles < needed_cycles{
        ic_cdk::eprintln!("Not enough cycles provided for sha256 operation.");
        return String::from("Not enough cycles provided for sha256 operation.");
    }
    let _ = msg_cycles_accept(needed_cycles);
    let mut input = msg.clone().into_bytes();
    ic_cdk::println!("input: {}", msg);
    for _ in 0..n {
        let mut hasher = Sha256::new();
        hasher.update(input);
        input = hasher.finalize().to_vec();
    }
    ic_cdk::println!("result: {}", hex::encode(input.clone()));
    hex::encode(input)
}

#[update]
pub fn sha256(msg :String, n :u32) -> String {
    let mut input = msg.clone().into_bytes();
    ic_cdk::println!("input: {}", msg);
    for _ in 0..n {
        let mut hasher = Sha256::new();
        hasher.update(input);
        input = hasher.finalize().to_vec();
    }
    ic_cdk::println!("result: {}", hex::encode(input.clone()));
    hex::encode(input)
}

#[query]
pub fn concat(s1:String, s2:String) -> String {
    format!("{} {}", s1, s2) 
}

#[update]
pub fn concat_with_cycles(s1:String, s2:String) -> String {
    let available_cycles = msg_cycles_available();
    let needed_cycles:u64 = CONCAT_STRING_COST_CYCLES;
    if available_cycles < needed_cycles{
        ic_cdk::eprintln!("Not enough cycles provided for concat operation.");
        return String::from("Not enough cycles provided for concat operation.");
    }
    let _ = msg_cycles_accept(needed_cycles);
    format!("{} {}", s1, s2) 
}
```
## Tips
- export identity .pem
  ```
  dfx identity list
  dfx identity use xxx // one of the listed idnetity
  dfx identity export xxx 
  ```
  
- get the identity's wallet
  ```
  dfx identity get-wallet
  ```
  
