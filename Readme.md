## Tetris

An implementation of the classic Tetris game, originally created by Alexey Pajitnov. 
This version intentionally mirrors the visual style of the original 1984 release, using bracket-based graphics with minor variations.

### Dev

See [roadmap.md]

Run
```
# Terminal 1:
go run main.go 2> tmp.log

# Terminal 2:
echo "" > tmp.log && tail -f tmp.log
```

Run tests:
```
go test
```
