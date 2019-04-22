

 ### General concept
For general concept explanation one can refer to /pkg/core/database/README.md. This document must focus on decisions made with regard to goleveldb specifics


### K/V storage schema to store a single `pkg/core/block.Block`

|    Prefix   | KEY                | VALUE                    | Count per block            |  Used by                 |
| :-----:     | :----------------: | :---------------------:  | :----------------------:   |:----------------------:  |
|  0x01       | HeaderHash         | Header.Encode()          | 1                          | 
|  0x02       | HeaderHash + TxID  | TxIndex + Tx.Encode()    | block txs count            | 
|  0x04       | TxID               | HeaderHash               | block txs count            | FetchBlockTxByHash
|  0x05       | KeyImage           | TxID                     | sum of block txs inputs    | FetchKeyImageExists
|  0x03       | Height             | HeaderHash               | 1                          | FetchBlockHashByHeight


Table notation
- HeaderHash - a calculated hash of block header
- TxID - a calculated hash of transaction
- \'+' operation - concatenation of byte arrays
- Tx.Encode() - Encoded binary form of all Tx fields without TxID