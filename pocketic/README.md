# PocketIC Golang: A Canister Testing Library

The client requires at least version 4 of the PocketIC server.
The client is not yet stable and is subject to change.

## References

- [PocketIC](https://github.com/dfinity/pocketic)
- [PocketIC Server](https://github.com/dfinity/ic/tree/master/rs/pocket_ic_server)

## List of Supported Endpoints

| Supported | Method | Endpoint                                          |
|-----------|--------|---------------------------------------------------|
| ✅         | GET    | /status                                           |
| ✅         | POST   | /blobstore                                        |
| ✅         | GET    | /blobstore/{id}                                   |
| ✅         | POST   | /verify_signature                                 |
| ✳️        | GET    | /read_graph/{state_label}/{op_id}                 |
| ✅         | GET    | /instances/                                       |
| ✅         | POST   | /instances/                                       |
| ✅         | DELETE | /instances/{id}                                   |
| ✅         | POST   | /instances/{id}/read/query                        |
| ✅         | GET    | /instances/{id}/read/get_time                     |
| ✅         | POST   | /instances/{id}/read/get_cycles                   |
| ✅         | POST   | /instances/{id}/read/get_stable_memory            |
| ✅         | POST   | /instances/{id}/read/get_subnet                   |
| ✅         | POST   | /instances/{id}/read/pub_key                      |
| ✅         | POST   | /instances/{id}/update/submit_ingress_message     |
| ✅         | POST   | /instances/{id}/update/await_ingress_message      |
| ✅         | POST   | /instances/{id}/update/execute_ingress_message    |
| ✅         | POST   | /instances/{id}/update/set_time                   |
| ✅         | POST   | /instances/{id}/update/add_cycles                 |
| ✅         | POST   | /instances/{id}/update/set_stable_memory          |
| ✅         | POST   | /instances/{id}/update/tick                       |
| ❌         | GET    | /instances/{id}/api/v2/status                     |
| ❌         | POST   | /instances/{id}/api/v2/canister/{ecid}/call       |
| ❌         | POST   | /instances/{id}/api/v2/canister/{ecid}/query      |
| ❌         | POST   | /instances/{id}/api/v2/canister/{ecid}/read_state |
| ✅         | POST   | /instances/{id}/auto_progress                     |
| ✅         | POST   | /instances/{id}/stop_progress                     |
| ✅         | POST   | /http_gateway/                                    |
| ✅         | POST   | /http_gateway/{id}/stop                           |


