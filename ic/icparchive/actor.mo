// Do NOT edit this file. It was automatically generated by https://github.com/aviate-labs/agent-go.
import T "types";

import { principalOfBlob } = "mo:⛔";

actor class _icparchive() : async actor {} {
    public query func get_blocks(_arg0 : T.GetBlocksArgs) : async (T.GetBlocksResult) {
        (#Err(#BadFirstBlockIndex({ requested_index = 6197789885467376506; first_valid_index = 11539667703195591131 })))
    };
    public query func get_encoded_blocks(_arg0 : T.GetBlocksArgs) : async (T.GetEncodedBlocksResult) {
        (#Err(#BadFirstBlockIndex({ requested_index = 6197789885467376506; first_valid_index = 11539667703195591131 })))
    };
}
