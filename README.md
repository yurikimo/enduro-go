Here’s a clean, **teaching-focused README.md** you can drop directly into your repo:

---

# 🏁 Enduro GO

A small **retro-style racing game** inspired by *Atari Enduro*, built with **Go** and **Ebiten**.

This project is designed as a **learning tool** — not just a game — to help developers understand how to build games in Go in a simple, practical, and fun way.

---

## 📷 Enduro - GO

Add your GIF here:

```md
![Demo](enduro.gif)
```

---

## 🎯 Goal of the Project

The main purpose of this project is to teach:

* How to structure a game in Go
* How a **game loop** works (`Update` / `Draw`)
* Basic **2D rendering with Ebiten**
* Simple **perspective and scaling**
* Clean and readable **Go code organization**

👉 The focus is **clarity over complexity**

---

## 🧠 What You Will Learn

### Game Architecture

* Separation of concerns:

  * `Game` → orchestrates everything
  * `Player` → input + movement
  * `Enemy` → AI + perspective behavior
  * `Road` → rendering + math
* State management:

  * Start screen
  * Running
  * Paused
  * Game Over

---

### Ebiten Game Loop

```go
func (g *Game) Update() error
func (g *Game) Draw(screen *ebiten.Image)
```

* `Update()` → game logic (movement, collisions, state)
* `Draw()` → rendering only

👉 **Rule:** Never mix logic and rendering

---

### Perspective (Fake 3D Effect)

The game simulates depth using scaling:

* Objects near the horizon → small
* Objects near the player → large

Key idea:

```go
progress := (y - horizon) / (screenHeight - horizon)
```

Used in:

* Enemy size
* Road width
* Positioning

---

### World Space vs Screen Space

* **World space** → logical position (lane, distance)
* **Screen space** → pixels drawn on screen

Example:

```go
options.GeoM.Scale(width/carSpriteWidth, height/carSpriteHeight)
options.GeoM.Translate(x, y)
```

👉 Sprites are **scaled to match perspective**

---

### Why `sync.Once` is used

```go
var spriteOnce sync.Once
```

* Ensures sprites are created **only once**
* Avoids:

  * unnecessary allocations
  * duplicated textures

👉 Important for performance and memory

---

### Value vs Pointer Receivers

You’ll see both used intentionally:

#### Value (`Player`, `Enemy` returned by value)

* Small structs
* Easy to copy
* Safer for beginners

#### Pointer (`Update`, mutation methods)

```go
func (e *Enemy) Update(...)
```

* Needed when modifying state

👉 Rule of thumb:

* **Read-only → value**
* **Modify → pointer**

---

## 🎮 Controls

| Key   | Action     |
| ----- | ---------- |
| ← / → | Steer      |
| ↑     | Accelerate |
| ↓     | Brake      |
| P     | Pause      |
| SPACE | Start game |
| R     | Restart    |

---

## 🚀 How to Run

### 1. Install Go

[https://go.dev/dl/](https://go.dev/dl/)

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Run the game

```bash
go run .
```

---

## 🗂 Project Structure

```
main.go           → Game loop & orchestration
player.go         → Player movement & input
enemy.go          → Enemy behavior & perspective
road.go           → Road rendering & math
sprites.go        → Sprite generation
sound.go          → Audio system
scoreManager.go   → Score persistence
utils.go          → Helper functions
```

---

## ⚠️ Important Design Choices

This project intentionally avoids:

* Over-engineering
* Complex abstractions
* Premature optimization

Instead, it favors:

* Simplicity
* Readability
* Learnability

---

## 🧪 Suggested Exercises

If you're learning, try:

* Add new enemy types
* Change lane system (more lanes)
* Add obstacles
* Improve collision detection
* Replace generated sprites with images
* Add UI with Ebiten text

---

## 💡 Why This Project Exists

Most Go tutorials focus on backend or CLI tools.

This project shows that:

👉 **Go can be used to build games in a clean and enjoyable way**

---

## 📌 Notes

* Not production-ready (by design)
* Built for learning and experimentation
* Keep things simple before making them complex

---

## 🙌 Contributions

Feel free to fork and experiment.

If you improve clarity or teaching value, even better.

---

## 🧑‍💻 Author

Built as a learning-focused project to explore **Go + game development**.

---
