## Tetris

### Dev

Run
```
# Terminal 1:
go run main.go 2> tmp.log

# Terminal 2:
echo "" > tmp.log && tail -f tmp.log
```

Run tests:
```
go test main.go main_test.go
```
