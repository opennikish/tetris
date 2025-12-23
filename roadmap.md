### Roadmap

#### Phase 1: PoC and Skeleton
- [x] UI framework (loop, user input handling, rendering)
- [x] PoC with a single tetromino falling down and being cemented
- [x] Cursor-based re-rendering
- [x] Add scenario test skeleton
- [x] Write scenario tests

#### Phase 2: Core Game Logic
- [ ] Introduce remaining tetrominos
- [ ] Fix known bugs
- [ ] Stabilize game abstractions
    - [ ] Re-rendering responsibility for tetrominos
    - [ ] Render test hooks — either implement them within tests (e.g., using a decorator around the `ScreenBuffer` test helper) or expose them via the app API
    - [ ] Tetromino representation — can a single struct support all tetrominos and their functionality (wall kicks, etc.)?
- [ ] Unit tests
- [ ] Side kicks
- [ ] Support rotations on the ground
    - [ ] Choose a standard
- [ ] Colors?

#### Phase 3:
- [ ] Add scoring system
    - [ ] Score
    - [ ] Total completed lines
    - [ ] Level
- [ ] Accelerate gravity
- [ ] Add a left-side panel with the next tetromino and statistics
- [ ] Add a right-side panel with control help

#### Phase 4
- [ ] Support windows
- [ ] Register in package managers
    - [ ] brew
    - [ ] debian (apt)
    - [ ] chocolatey