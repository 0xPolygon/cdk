docker run --rm -v $(pwd):/contracts ethereum/solc:0.8.18-alpine - /contracts/verifybatchesmock/VerifyBatchesMock.sol -o /contracts --abi --bin --overwrite --optimize
mv -f VerifyBatchesMock.abi abi/verifybatchesmock.abi
mv -f VerifyBatchesMock.bin bin/verifybatchesmock.bin
rm -f IBasePolygonZkEVMGlobalExitRoot.abi
rm -f IBasePolygonZkEVMGlobalExitRoot.bin
rm -f IPolygonZkEVMGlobalExitRootV2.abi
rm -f IPolygonZkEVMGlobalExitRootV2.bin

docker run --rm -v $(pwd):/contracts ethereum/solc:0.8.18-alpine - /contracts/claimmock/ClaimMock.sol -o /contracts --abi --bin --overwrite --optimize
mv -f ClaimMock.abi abi/claimmock.abi
mv -f ClaimMock.bin bin/claimmock.bin

docker run --rm -v $(pwd):/contracts ethereum/solc:0.8.18-alpine - /contracts/claimmockcaller/ClaimMockCaller.sol -o /contracts --abi --bin --overwrite --optimize
mv -f ClaimMockCaller.abi abi/claimmockcaller.abi
mv -f ClaimMockCaller.bin bin/claimmockcaller.bin
rm -f IClaimMock.abi
rm -f IClaimMock.bin