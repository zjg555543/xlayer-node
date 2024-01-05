package config

// DefaultValues is the default configuration
const DefaultValues = `
Port = 8080

[L1]
ChainId = 1337
RPC = "http://127.0.0.1:8545"
PolygonZkEVMAddress = "0x0D9088C72Cd4F08e9dDe474D8F5394147f64b22C"
SeqPrivateKey = {Path = "/pk/sequencer.keystore", Password = "testonly"}
AggPrivateKey = {Path = "/pk/aggregator.keystore", Password = "testonly"}


[Log]
Environment = "development"
Level = "error"
Outputs = ["stdout"]
`
