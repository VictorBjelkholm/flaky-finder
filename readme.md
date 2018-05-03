## flaky-finder
> Jenkins util that finds the tests that are most failing, given an job URL

### Usage

Outputs the top 10 failing test-cases

```
go get -v ./...
go run main.go https://ci.ipfs.team/job/IPFS/job/js-ipfs-api/job/master/
```
