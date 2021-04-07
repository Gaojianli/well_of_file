all: proto

proto:
	protoc --gofast_out=protobuf idl/protocol.proto

aar:
	gomobile bind -o output/WellOfFile.aar -target=android github.com/Gaojianli/well_of_file