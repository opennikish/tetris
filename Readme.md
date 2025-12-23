## Tetris

An implementation of the Tetris game, originally created by Alexey Pajitnov.
This version intentionally mirrors the visual style and esthetic of the original 1984 release, using bracket-based graphics.

Intentionaly has zero dependencies.

### Dev

See [roadmap.md](./roadmap.md)

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
