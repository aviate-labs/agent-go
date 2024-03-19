// Do NOT edit this file. It was automatically generated by https://github.com/aviate-labs/agent-go.
import T "types";

import { principalOfBlob } = "mo:⛔";

actor class _wallet() : async actor {} {
    public query func wallet_api_version() : async (Text) {
        ("12501942321308663592")
    };
    public query func name() : async (?Text) {
        (?"12501942321308663592")
    };
    public shared func set_name(_arg0 : Text) : async () {
        ()
    };
    public query func get_controllers() : async ([Principal]) {
        ([ principalOfBlob("xAABD0E49110AC27F321B5538A4BD8E1D253F8AFFD1A9C81C5B6EBD2C7EE44EF3"), principalOfBlob("x8A251E9C55DCC87B1BCA3A098396CB5C620BF809689EE32C89B845322F01379C"), principalOfBlob("x4C41BFC610298858C4967345654F9614DC1B2711874D3E6FE25E3A8F72C632E3"), principalOfBlob("x5E92AF8BF9CEF5969C068EE637EBB34E1E18DC1513DD22FC38BCA3DEA73A02A6") ])
    };
    public shared func add_controller(_arg0 : Principal) : async () {
        ()
    };
    public shared func remove_controller(_arg0 : Principal) : async (T.WalletResult) {
        (#Ok(()))
    };
    public query func get_custodians() : async ([Principal]) {
        ([ principalOfBlob("xAABD0E49110AC27F321B5538A4BD8E1D253F8AFFD1A9C81C5B6EBD2C7EE44EF3"), principalOfBlob("x8A251E9C55DCC87B1BCA3A098396CB5C620BF809689EE32C89B845322F01379C"), principalOfBlob("x4C41BFC610298858C4967345654F9614DC1B2711874D3E6FE25E3A8F72C632E3"), principalOfBlob("x5E92AF8BF9CEF5969C068EE637EBB34E1E18DC1513DD22FC38BCA3DEA73A02A6") ])
    };
    public shared func authorize(_arg0 : Principal) : async () {
        ()
    };
    public shared func deauthorize(_arg0 : Principal) : async (T.WalletResult) {
        (#Ok(()))
    };
    public query func wallet_balance() : async ({ amount : Nat64 }) {
        ({ amount = 12501942321308663592 })
    };
    public query func wallet_balance128() : async ({ amount : Nat }) {
        ({ amount = 12501942321308663592 })
    };
    public shared func wallet_send(_arg0 : { canister : Principal; amount : Nat64 }) : async (T.WalletResult) {
        (#Ok(()))
    };
    public shared func wallet_send128(_arg0 : { canister : Principal; amount : Nat }) : async (T.WalletResult) {
        (#Ok(()))
    };
    public shared func wallet_receive(_arg0 : ?T.ReceiveOptions) : async () {
        ()
    };
    public shared func wallet_create_canister(_arg0 : T.CreateCanisterArgs) : async (T.WalletResultCreate) {
        (#Ok({ canister_id = principalOfBlob("xAABD0E49110AC27F321B5538A4BD8E1D253F8AFFD1A9C81C5B6EBD2C7EE44EF3") }))
    };
    public shared func wallet_create_canister128(_arg0 : T.CreateCanisterArgs128) : async (T.WalletResultCreate) {
        (#Ok({ canister_id = principalOfBlob("xAABD0E49110AC27F321B5538A4BD8E1D253F8AFFD1A9C81C5B6EBD2C7EE44EF3") }))
    };
    public shared func wallet_create_wallet(_arg0 : T.CreateCanisterArgs) : async (T.WalletResultCreate) {
        (#Ok({ canister_id = principalOfBlob("xAABD0E49110AC27F321B5538A4BD8E1D253F8AFFD1A9C81C5B6EBD2C7EE44EF3") }))
    };
    public shared func wallet_create_wallet128(_arg0 : T.CreateCanisterArgs128) : async (T.WalletResultCreate) {
        (#Ok({ canister_id = principalOfBlob("xAABD0E49110AC27F321B5538A4BD8E1D253F8AFFD1A9C81C5B6EBD2C7EE44EF3") }))
    };
    public shared func wallet_store_wallet_wasm(_arg0 : { wasm_module : Blob }) : async () {
        ()
    };
    public shared func wallet_call(_arg0 : { canister : Principal; method_name : Text; args : Blob; cycles : Nat64 }) : async (T.WalletResultCall) {
        (#Ok({ return = "xAABD0E49110AC27F321B5538A4BD8E1D253F8AFFD1A9C81C5B6EBD2C7EE44EF3" }))
    };
    public shared func wallet_call128(_arg0 : { canister : Principal; method_name : Text; args : Blob; cycles : Nat }) : async (T.WalletResultCall) {
        (#Ok({ return = "xAABD0E49110AC27F321B5538A4BD8E1D253F8AFFD1A9C81C5B6EBD2C7EE44EF3" }))
    };
    public shared func add_address(_arg0 : T.AddressEntry) : async () {
        ()
    };
    public query func list_addresses() : async ([T.AddressEntry]) {
        ([ { id = principalOfBlob("xAABD0E49110AC27F321B5538A4BD8E1D253F8AFFD1A9C81C5B6EBD2C7EE44EF3"); name = ?"9207851698408531338"; kind = #Canister; role = #Contact }, { id = principalOfBlob("x9EE32C89B845322F01379CB07CF44C41BFC610298858C4967345654F9614DC1B"); name = ?"18342943986503597645"; kind = #User; role = #Custodian }, { id = principalOfBlob("x969C068EE637EBB34E1E18DC1513DD22FC38BCA3DEA73A02A6B18EB21D8C751B"); name = ?"15380023008932546055"; kind = #User; role = #Controller }, { id = principalOfBlob("x39BC49BC882A10B6BBF3F350B6D2D41664A1B826BBE574D8BA20F61084C26198"); name = ?"11315935127771919894"; kind = #Unknown; role = #Controller } ])
    };
    public shared func remove_address(_arg0 : Principal) : async (T.WalletResult) {
        (#Ok(()))
    };
    public query func get_events(_arg0 : ?{ from : ?Nat32; to : ?Nat32 }) : async ([T.Event]) {
        ([ { id = 2949695419; timestamp = 10285557685837050495; kind = #CyclesSent({ to = principalOfBlob("xA9C81C5B6EBD2C7EE44EF3CD9E1B8A251E9C55DCC87B1BCA3A098396CB5C620B"); amount = 12408056582236857246; refund = 8643670682807763247 }) }, { id = 1447520861; timestamp = 15226500214129345624; kind = #CanisterCalled({ canister = principalOfBlob("x4D3E6FE25E3A8F72C632E3513A5F5E92AF8BF9CEF5969C068EE637EBB34E1E18"); method_name = "5539044620517515997"; cycles = 1923756884486142631 }) }, { id = 2805007336; timestamp = 15380023008932546055; kind = #CyclesReceived({ from = principalOfBlob("x70087656C06A2339BC49BC882A10B6BBF3F350B6D2D41664A1B826BBE574D8BA"); amount = 7929607626476864132; memo = ?"11315935127771919894" }) }, { id = 1415628053; timestamp = 13683296349367290469; kind = #WalletDeployed({ canister = principalOfBlob("xF6FDE26C9239614A5D20BE64DB35184AFC8405CAAA7A74B6930B2F2C16BB7A5A") }) } ])
    };
    public query func get_events128(_arg0 : ?{ from : ?Nat32; to : ?Nat32 }) : async ([T.Event128]) {
        ([ { id = 2949695419; timestamp = 10285557685837050495; kind = #CyclesSent({ to = principalOfBlob("xA9C81C5B6EBD2C7EE44EF3CD9E1B8A251E9C55DCC87B1BCA3A098396CB5C620B"); amount = 12408056582236857246; refund = 8643670682807763247 }) }, { id = 1447520861; timestamp = 15226500214129345624; kind = #CanisterCalled({ canister = principalOfBlob("x4D3E6FE25E3A8F72C632E3513A5F5E92AF8BF9CEF5969C068EE637EBB34E1E18"); method_name = "5539044620517515997"; cycles = 1923756884486142631 }) }, { id = 2805007336; timestamp = 15380023008932546055; kind = #CyclesReceived({ from = principalOfBlob("x70087656C06A2339BC49BC882A10B6BBF3F350B6D2D41664A1B826BBE574D8BA"); amount = 7929607626476864132; memo = ?"11315935127771919894" }) }, { id = 1415628053; timestamp = 13683296349367290469; kind = #WalletDeployed({ canister = principalOfBlob("xF6FDE26C9239614A5D20BE64DB35184AFC8405CAAA7A74B6930B2F2C16BB7A5A") }) } ])
    };
    public query func get_chart(_arg0 : ?{ count : ?Nat32; precision : ?Nat64 }) : async ([(Nat64, Nat64)]) {
        ([ ( 7404491806706941354, 10285557685837050495 ), ( 16344125491934207374, 6353661455985592489 ), ( 13842832487040869502, 9207851698408531338 ), ( 7031951943849876347, 10982038652290489547 ) ])
    };
    public query func list_managed_canisters(_arg0 : { from : ?Nat32; to : ?Nat32 }) : async ([T.ManagedCanisterInfo], Nat32) {
        ([ { id = principalOfBlob("xAABD0E49110AC27F321B5538A4BD8E1D253F8AFFD1A9C81C5B6EBD2C7EE44EF3"); name = ?"9207851698408531338"; created_at = 7031951943849876347 }, { id = principalOfBlob("xCB5C620BF809689EE32C89B845322F01379CB07CF44C41BFC610298858C49673"); name = ?"8829044454151951510"; created_at = 18342943986503597645 }, { id = principalOfBlob("x72C632E3513A5F5E92AF8BF9CEF5969C068EE637EBB34E1E18DC1513DD22FC38"); name = ?"1923756884486142631"; created_at = 10068773106139630621 }, { id = principalOfBlob("x0746C29470D17024B544336AECB870087656C06A2339BC49BC882A10B6BBF3F3"); name = ?"10428971936530044628"; created_at = 7786994376157721829 } ], 107831972)
    };
    public query func get_managed_canister_events(_arg0 : { canister : Principal; from : ?Nat32; to : ?Nat32 }) : async (?[T.ManagedCanisterEvent]) {
        (?[ { id = 2949695419; timestamp = 10285557685837050495; kind = #Created({ cycles = 6353661455985592489 }) }, { id = 3010102092; timestamp = 9207851698408531338; kind = #Created({ cycles = 10982038652290489547 }) }, { id = 895428951; timestamp = 8643670682807763247; kind = #CyclesSent({ amount = 15226500214129345624; refund = 8829044454151951510 }) }, { id = 3774773420; timestamp = 2188532067303868018; kind = #Called({ method_name = "4389663725167484054"; cycles = 9589032066643545779 }) } ])
    };
    public query func get_managed_canister_events128(_arg0 : { canister : Principal; from : ?Nat32; to : ?Nat32 }) : async (?[T.ManagedCanisterEvent128]) {
        (?[ { id = 2949695419; timestamp = 10285557685837050495; kind = #Created({ cycles = 6353661455985592489 }) }, { id = 3010102092; timestamp = 9207851698408531338; kind = #Created({ cycles = 10982038652290489547 }) }, { id = 895428951; timestamp = 8643670682807763247; kind = #CyclesSent({ amount = 15226500214129345624; refund = 8829044454151951510 }) }, { id = 3774773420; timestamp = 2188532067303868018; kind = #Called({ method_name = "4389663725167484054"; cycles = 9589032066643545779 }) } ])
    };
    public shared func set_short_name(_arg0 : Principal, _arg1 : ?Text) : async (?T.ManagedCanisterInfo) {
        (?{ id = principalOfBlob("x287F06984DD27FAABD0E49110AC27F321B5538A4BD8E1D253F8AFFD1A9C81C5B"); name = ?"13842832487040869502"; created_at = 9207851698408531338 })
    };
    public query func http_request(_arg0 : T.HttpRequest) : async (T.HttpResponse) {
        ({ status_code = 38652; headers = [ ( "10285557685837050495", "16344125491934207374" ), ( "6353661455985592489", "13842832487040869502" ), ( "9207851698408531338", "7031951943849876347" ), ( "10982038652290489547", "12408056582236857246" ) ]; body = "x2F01379CB07CF44C41BFC610298858C4967345654F9614DC1B2711874D3E6FE2"; streaming_strategy = ?#Callback({ callback = { /* func */ }; token = {  } }) })
    };
}
