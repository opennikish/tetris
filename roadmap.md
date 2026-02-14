### Roadmap

#### Phase 1: PoC and Skeleton
- [x] UI framework (loop, user input handling, rendering)
- [x] PoC with a single tetromino falling down and being cemented
- [x] Cursor-based re-rendering
- [x] Add scenario test skeleton
- [x] Write scenario tests

#### Phase 2: Core Game Logic
- [x] Separate core game logic from presentation
- [x] Restructure: split `main.go` into packages
- [x] Introduce remaining tetrominos
- [x] Add hard drop
- [x] Wipe out copmpleted rows
- [x] Fix known bugs
- [x] Stabilize game abstractions    
    - [ ] Render test hooks 
        — either implement them within tests (e.g., using a decorator around the `ScreenBuffer` test helper) or expose them via the app API
    - [x] Tetromino representation — can a single struct support all tetrominos and their functionality (wall kicks, etc.)?
- [ ] Unit tests
- [x] Choose rotation system standard    
- [ ] Center playfield
- [x] Refine quit, interrupt and gameover events handling
- [ ] Gameplay screen
    - [ ] Add a left-side panel 
        - [ ] Next tetromino
        - [ ] Stats
            - [ ] Score
            - [ ] Total completed lines
            - [ ] Level
    - [ ] Add a right-side panel with control help
- [ ] Rendering delay
- [ ] Set cursor to 0,0 after render

#### Phase 3:
- [ ] Add clocks header       
- [ ] Color: native, green
- [ ] Allow choose blocks instead of brakets
- [ ] Engine
    - [ ] Screens update                
    - [ ] Extract loop
    - [ ] Update loop frequency (for level changing)
- [ ] Add remaining screens
    - [ ] Entry screen (choose level)
    - [ ] Leaderboard        
- [ ] Terminal:
    - [ ] open new tty like nano
    - [ ] screen width (for centering)
    - [ ] multiplatform
- [ ] ?? Accelerate gravity (note 1st version did't have it, only second)

#### Phase 4
- [ ] Update readme
    - [ ] Architecture design
    - [ ] References (standarts, videos, etc)
- [ ] Register in package managers
    - [ ] brew
    - [ ] debian (apt)
    - [ ] chocolatey
- [ ] ?? Debug mode which saves user actions and state to file