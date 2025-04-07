package main

import (
	"image/color"
	"log"
	"math"

	// "strconv"

	"github.com/hajimehoshi/ebiten/v2"
	// "github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	// "golang.org/x/image/font/basicfont"
)

var outlineColor = color.RGBA{0, 0, 0, 89}

const blockSize float32 = 60

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
	Shape    Shape
	Selected bool
}

type World struct {
	Field        [][]Block
	Score        int
	LastInputs   [][2]int
	RandomBlocks []Block
	TimeElapsed  float64
	Changed      bool
	Combo        int
}

type Game struct {
	world World
}

func DefBlock(s Shape) Block {
	return Block{
		Shape:    s,
		Selected: false,
	}
}

func getBlockAt(x, y int, w World) *Block {
	if y < 0 || y >= len(w.Field) || x < 0 || x >= len(w.Field[y]) {
		return nil
	}
	return &w.Field[x][y]
}
func unSelectAllWorld(w *World) {
	for _, row := range w.Field {
		for i := range row {
			row[i].Selected = false
		}
	}
}

func selectBlockWorld(x, y int, w *World) {
	unSelectAllWorld(w)
	if b := getBlockAt(x, y, *w); b != nil {
		b.Selected = true
	}
}

func swap(pos1, pos2 [2]int, w *World) {
	x1, y1 := pos1[0], pos1[1]
	x2, y2 := pos2[0], pos2[1]
	if b1 := getBlockAt(x1, y1, *w); b1 != nil {
		if b2 := getBlockAt(x2, y2, *w); b2 != nil {
			w.Field[y1][x1], w.Field[y2][x2] = *b2, *b1
		}
	}
}

func connectedHelper(fld [][]Block, pos [2]int, visited map[[2]int]bool) map[[2]int]bool {
	x, y := pos[0], pos[1]
	if x < 0 || x >= len(fld[0]) || y < 0 || y >= len(fld) || visited[pos] {
		return visited
	}
	visited[pos] = true
	currentShape := fld[y][x].Shape
	neighbors := [][2]int{{x - 1, y}, {x + 1, y}, {x, y - 1}, {x, y + 1}}
	for _, neighbor := range neighbors {
		nx, ny := neighbor[0], neighbor[1]
		if nx >= 0 && nx < len(fld[0]) && ny >= 0 && ny < len(fld) {
			if fld[ny][nx].Shape == currentShape {
				visited = connectedHelper(fld, neighbor, visited)
			}
		}
	}
	return visited
}

func getSelectedBlock(w *World) [2]int {
	for y, row := range w.Field {
		for x, b := range row {
			if b.Selected {
				return [2]int{x, y}
			}
		}
	}
	return [2]int{-1, -1}
}

func neighbors(pos1, pos2 [2]int) bool {
	dx := pos1[0] - pos2[0]
	dy := pos1[1] - pos2[1]
	return (dx == 1 && dy == 0) || (dx == -1 && dy == 0) || (dx == 0 && dy == 1) || (dx == 0 && dy == -1)
}

func handleSelect(x, y int, w *World) {
	if w.Changed {
		return
	}
	b := getBlockAt(x, y, *w)
	if b == nil {
		return
	}
	if b.Selected {
		unSelectAllWorld(w)
	} else {
		selectedBlock := getSelectedBlock(w)
		if selectedBlock != [2]int{-1, -1} {
			sx, sy := selectedBlock[0], selectedBlock[1]
			if neighbors([2]int{x, y}, [2]int{sx, sy}) {
				swap([2]int{x, y}, [2]int{sx, sy}, w)
				w.TimeElapsed = 0
				w.Changed = true
				del1 := connectedHelper(w.Field, [2]int{x, y}, make(map[[2]int]bool))
				del2 := connectedHelper(w.Field, [2]int{sx, sy}, make(map[[2]int]bool))
				if len(del1) >= 4 || len(del2) >= 4 {
					unSelectAllWorld(w)
				} else {
					selectBlockWorld(x, y, w)
				}
			} else {
				selectBlockWorld(x, y, w)
			}
		} else {
			selectBlockWorld(x, y, w)
		}
	}
}

func getAllDeletions(fld [][]Block) [][2]int {
	visited := make(map[[2]int]bool)
	var deletions [][2]int
	for y, row := range fld {
		for x := range row {
			if !visited[[2]int{x, y}] {
				connected := connectedHelper(fld, [2]int{x, y}, make(map[[2]int]bool))
				if len(connected) >= 4 {
					for pos := range connected {
						deletions = append(deletions, pos)
					}
				}
				for pos := range connected {
					visited[pos] = true
				}
			}
		}
	}
	return deletions
}

func processDeletions(coords [][2]int, w *World) {
	scorePlus := 5 * w.Combo * len(coords) * int(math.Log2(float64(len(coords))))
	w.Score += scorePlus
	for _, coord := range coords {
		x, y := coord[0], coord[1]
		w.Field[y][x] = DefBlock(Blank)
	}
}

func containsBlank(fld [][]Block) bool {
	for _, row := range fld {
		for _, b := range row {
			if b.Shape == Blank {
				return true
			}
		}
	}
	return false
}

func updateOne(w *World, x int) {
	for y := len(w.Field) - 1; y >= 0; y-- {
		if w.Field[y][x].Shape == Blank {
			for ny := y - 1; ny >= 0; ny-- {
				if w.Field[ny][x].Shape != Blank {
					w.Field[y][x], w.Field[ny][x] = w.Field[ny][x], w.Field[y][x]
					break
				}
			}
		}
	}
}

func pointToCoords(mx, my int) (int, int) {
	x := mx / int(blockSize)
	y := my / int(blockSize)
	return x, y
}

type Point struct {
	X, Y float32
}

func scalePoints(pts []Point, sx, sy float32) []Point {
	result := make([]Point, len(pts))
	for i, p := range pts {
		result[i] = Point{p.X * sx, p.Y * sy}
	}
	return result
}

func translatePoints(pts []Point, tx, ty float32) []Point {
	result := make([]Point, len(pts))
	for i, p := range pts {
		result[i] = Point{p.X + tx, p.Y + ty}
	}
	return result
}

func scaleGemDownPoints(pts []Point) []Point {
	scaled := scalePoints(pts, 0.96, 0.96)
	translated := translatePoints(scaled, 0.02, 0.02)
	return translated
}

func getBasePoints(s Shape) []Point {
	switch s {
	case Blank:
		return []Point{{0, 0}, {1, 0}, {1, 1}, {0, 1}}
	case Triangle:
		return []Point{{0.1, 0.17}, {0.5, 0}, {0.9, 0.17}, {0.5, 1.4}}
	case Square:
		return []Point{{0, 0}, {1, 0}, {1, 1}, {0, 1}}
	case Octa:
		return []Point{{0.25, 0}, {0.75, 0}, {1, 0.25}, {1, 0.75}, {0.75, 1}, {0.25, 1}, {0, 0.75}, {0, 0.25}}
	case Pent:
		return []Point{{0.15, 0}, {0.85, 0}, {1, 0.6}, {0.5, 1.1}, {0, 0.6}}
	case Diamond:
		return []Point{{0.5, 0}, {0.83, 0.5}, {0.5, 1}, {0.17, 0.5}}
	case Hexa:
		return []Point{{0, 0.5}, {0.26, 0.95}, {0.74, 0.95}, {1, 0.5}, {0.74, 0.05}, {0.26, 0.05}}
	default:
		return []Point{}
	}
}

func getTransformedPoints(s Shape) []Point {
	basePts := getBasePoints(s)
	var transformedPts []Point
	switch s {
	case Triangle:
		transformedPts = scalePoints(basePts, 1, 1/1.4)
	case Pent:
		transformedPts = scalePoints(basePts, 1, 1/1.1)
	case Square:
		transformedPts = scaleGemDownPoints(basePts)
	default:
		transformedPts = basePts
	}
	finalPts := scaleGemDownPoints(transformedPts)
	return finalPts
}

func getShapeColor(s Shape) color.RGBA {
	switch s {
	case Blank:
		return color.RGBA{255, 255, 255, 255}
	case Triangle:
		return color.RGBA{139, 0, 139, 255}
	case Square:
		return color.RGBA{224, 191, 38, 255}
	case Octa:
		return color.RGBA{0, 127, 255, 255}
	case Pent:
		return color.RGBA{255, 140, 0, 255}
	case Diamond:
		return color.RGBA{169, 169, 169, 255}
	case Hexa:
		return color.RGBA{63, 127, 0, 255}
	default:
		return color.RGBA{0, 0, 0, 0}
	}
}

func dimColor(c color.RGBA) color.RGBA {
	return color.RGBA{
		R: uint8(float32(c.R) * 0.7),
		G: uint8(float32(c.G) * 0.7),
		B: uint8(float32(c.B) * 0.7),
		A: c.A,
	}
}

func pointsToVertices(pts []Point, blockX, blockY, blockSize float32, col color.RGBA) []ebiten.Vertex {
	vertices := make([]ebiten.Vertex, len(pts))
	for i, p := range pts {
		vertices[i] = ebiten.Vertex{
			DstX:   blockX + p.X*blockSize,
			DstY:   blockY + p.Y*blockSize,
			SrcX:   0,
			SrcY:   0,
			ColorR: float32(col.R) / 255,
			ColorG: float32(col.G) / 255,
			ColorB: float32(col.B) / 255,
			ColorA: float32(col.A) / 255,
		}
	}
	return vertices
}

func computeInnerPoints(pts []Point) []Point {
	inner := make([]Point, len(pts))
	for i, p := range pts {
		inner[i] = Point{
			X: (p.X + 0.5) / 2,
			Y: (p.Y + 0.5) / 2,
		}
	}
	return inner
}

func drawFilledPolygon(screen *ebiten.Image, vertices []ebiten.Vertex) {
	if len(vertices) < 3 {
		return
	}
	indices := make([]uint16, 0, 3*(len(vertices)-2))
	for i := 1; i < len(vertices)-1; i++ {
		indices = append(indices, 0, uint16(i), uint16(i+1))
	}
	screen.DrawTriangles(vertices, indices, nil, nil)
}

func drawOutline(screen *ebiten.Image, vertices []ebiten.Vertex) {
	for i := 0; i < len(vertices); i++ {
		p1 := vertices[i]
		p2 := vertices[(i+1)%len(vertices)]
		vector.StrokeLine(screen, p1.DstX, p1.DstY, p2.DstX, p2.DstY, 1, outlineColor, false)
	}
}

func drawGem(screen *ebiten.Image, s Shape, blockX, blockY, blockSize float32) {
	if s == Blank {
		return
	}

	pts := getTransformedPoints(s)
	col := getShapeColor(s)

	vertices := pointsToVertices(pts, blockX, blockY, blockSize, col)
	drawFilledPolygon(screen, vertices)
	drawOutline(screen, vertices)

	innerPts := computeInnerPoints(pts)
	innerCol := dimColor(col)
	innerVertices := pointsToVertices(innerPts, blockX, blockY, blockSize, innerCol)
	drawFilledPolygon(screen, innerVertices)
	drawOutline(screen, innerVertices)
}

func drawBlock(screen *ebiten.Image, b Block, gx, gy int, blockSize float32) {
	blockX := float32(gx) * blockSize
	blockY := float32(gy) * blockSize
	drawGem(screen, b.Shape, blockX, blockY, blockSize)
	if b.Selected {
		vector.StrokeRect(screen, blockX, blockY, blockSize, blockSize, 2, outlineColor, false)
	}
}

func drawField(screen *ebiten.Image, fld [][]Block, blockSize float32) {
	for y, row := range fld {
		for x, b := range row {
			drawBlock(screen, b, x, y, blockSize)
		}
	}
}

func (g *Game) Update() error {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		x, y := pointToCoords(mx, my)
		handleSelect(x, y, &g.world)
	}

	g.world.TimeElapsed += 1.0 / 60.0
	if g.world.TimeElapsed >= 0.5 {
		if containsBlank(g.world.Field) {
			for x := 0; x < len(g.world.Field[0]); x++ {
				updateOne(&g.world, x)
			}
			g.world.TimeElapsed = 0
		} else if g.world.Changed {
			deletions := getAllDeletions(g.world.Field)
			if len(deletions) > 0 {
				processDeletions(deletions, &g.world)
				g.world.TimeElapsed = 0
				g.world.Combo++
			} else {
				g.world.Changed = false
				g.world.Combo = 0
			}
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	drawField(screen, g.world.Field, blockSize)
	// text.Draw(screen, "Score: "+strconv.Itoa(g.world.Score), basicfont.Face7x13, 10, 20, color.Black)
	// text.Draw(screen, "Combo: "+strconv.Itoa(g.world.Combo), basicfont.Face7x13, 10, 40, color.Black)
	// text.Draw(screen, "Instructions: Match at least 4 blocks", basicfont.Face7x13, 10, 60, color.Black)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 480, 480
}

func main() {

	field := make([][]Block, 8)
	for i := range field {
		field[i] = make([]Block, 8)
		for j := range field[i] {
			shape := Shape(j % 6)
			field[i][j] = DefBlock(shape)
		}
	}

	game := &Game{
		world: World{
			Field: field,
			Score: 0,
		},
	}

	ebiten.SetWindowSize(480, 480)
	ebiten.SetWindowTitle("Go Tile Match Prototype")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
