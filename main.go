package main

import (
    "github.com/hajimehoshi/ebiten/v2"
    "log"
)

type Shape int

const (
    Triangle Shape = iota
    Square
    Octa
    Pent
    Diamond
    Hexa
    Blank
)

type Block struct {
    Shape Shape
    Selected bool
}

type World struct {
    Field [][]Block
    Score int
    LastInputs [][2]int
    RandomBlocks []Block
    TimeElapsed float64
    Changed bool
    Combo int
}

type Game struct {
    world World
}

func (g *Game) Update() error {
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    // TODO: Rendering logic
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return 480, 480
}


func main() {
    
    game := &Game{}

    ebiten.SetWindowSize(480, 480)
    ebiten.SetWindowTitle("Go Tile Match Prototype")

    if err := ebiten.RunGame(game); err != nil {
        log.Fatal(err)
    }
}