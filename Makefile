all: make_proto

make_proto:
	protoc --gofast_out=protobuf idl/protocol.proto