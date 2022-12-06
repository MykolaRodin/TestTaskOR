package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultFieldWidth        int = 8
	defaultFieldHeight       int = 8
	defaultBlackHoleQuantity int = 10
)

func main() {
	var help bool = false
	var fieldWidth int = defaultFieldWidth
	var fieldHeight int = defaultFieldHeight
	var blackHoleQuantity int = defaultBlackHoleQuantity

	flag.BoolVar(&help, "help", false, "Show help")
	flag.IntVar(&fieldWidth, "fieldWidth", defaultFieldWidth, "Field Width")
	flag.IntVar(&fieldHeight, "fieldHeight", defaultFieldHeight, "Field Height")
	flag.IntVar(&blackHoleQuantity, "blackHoleQuantity", defaultBlackHoleQuantity, "Black Hole Quantity")
	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(0)
	}

	// Initialize play field
	playField, err := NewPlayField(fieldWidth, fieldHeight, blackHoleQuantity)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(0)
	}
	playField.addBlackHoles()
	playField.addAdjacentBlackHoles()
	playField.printField()

	// Make endless loop to receive user commands. The application exits when the game is won or lost.
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$ ")
		commandParameters, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		width, height, marked, err := playField.parseUserCommand(commandParameters)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			playField.updateFieldCell(width, height, marked)

			if playField.checkFieldBlackHole(width, height) {
				// Force total visibility if Black Hole has been found
				for h := 0; h < playField.Width; h++ {
					for w := 0; w < playField.Width; w++ {
						playField.updateFieldCell(w, h, false)
					}
				}
				playField.printField()
				fmt.Fprintln(os.Stderr, "!!!!!!! BLACK HOLE HAS BEEN EXPLODED !!!!!!!")
				os.Exit(0)
			}

			if !marked {
				playField.updateFieldVisibility(width, height, map[playCellPos]struct{}{})
			}
			playField.printField()

			if playField.checkWinResult() {
				fmt.Fprintln(os.Stderr, "!!!!!!! YOU WON !!!!!!!")
				os.Exit(0)
			}
		}
	}
}

// The subsequent types and functions relate to play field are not placed in a separate file for simplicity
type playCellPos struct {
	width  int
	height int
}

type playCell struct {
	IsVisible                 bool
	IsBlackHole               bool
	IsMarked                  bool
	IsExploded                bool
	AdjucentBlackHoleQuantity int8
}

type playField struct {
	Width             int
	Height            int
	BlackHoleQuantity int
	Cells             []playCell
}

// NewPlayField creates new play field with all invisible cells and without black holes
func NewPlayField(fieldWidth, fieldHeight, blackHoleQuantity int) (*playField, error) {
	fmt.Fprintf(os.Stdout, "Creating field with fieldWidth=%d fieldHight=%d blackHoleQuantity=%d...\n",
		fieldWidth, fieldHeight, blackHoleQuantity)

	if fieldWidth <= 0 {
		return nil, errors.New("field width should be higher than 0")
	}
	if fieldHeight <= 0 {
		return nil, errors.New("field height should be higher than 0")
	}
	if blackHoleQuantity < 1 {
		return nil, errors.New("black hole quantity should be higher than 0")
	}
	if blackHoleQuantity > fieldWidth*fieldHeight {
		return nil, errors.New("black hole quantity should NOT be higher than the multiplication field width on field height")
	}

	cells := make([]playCell, fieldWidth*fieldHeight)
	field := &playField{
		Width:             fieldWidth,
		Height:            fieldHeight,
		BlackHoleQuantity: blackHoleQuantity,
		Cells:             cells,
	}

	fmt.Fprintln(os.Stdout, "please take into account in console version both width and haight are counted from top-left corner and are zero based")
	fmt.Fprintln(os.Stderr, "please enter 'width height (marked)'")

	return field, nil
}

// addBlackHoles randomly adds the required quantity of black holes
func (field *playField) addBlackHoles() {
	fieldSize := field.Width * field.Height
	attempts := 0
	maxAttempts := field.BlackHoleQuantity * 100
	blackHoles := 0
	for blackHoles < field.BlackHoleQuantity {
		rand.Seed(time.Now().UnixNano())
		blackHolePos := rand.Intn(fieldSize)
		if !field.Cells[blackHolePos].IsBlackHole {
			field.Cells[blackHolePos].IsBlackHole = true
			blackHoles++
			attempts++
		} else if attempts >= field.BlackHoleQuantity*100 {
			panic(fmt.Sprintf("Could not generate %d black holes in %d attempts",
				field.BlackHoleQuantity, maxAttempts))
		}
	}
}

// addAdjacentBlackHoles add the quantity of adjusting black holes to every cell in the field
func (field *playField) addAdjacentBlackHoles() {
	for height := 0; height < field.Height; height++ {
		for width := 0; width < field.Width; width++ {
			rowBias := height * field.Width
			currentPos := rowBias + width
			currentCell := &field.Cells[currentPos]
			// Left
			if width > 0 && field.Cells[currentPos-1].IsBlackHole {
				currentCell.AdjucentBlackHoleQuantity++
			}
			// Right
			if width < field.Width-1 && field.Cells[currentPos+1].IsBlackHole {
				currentCell.AdjucentBlackHoleQuantity++
			}
			// Up
			if height > 0 && field.Cells[currentPos-field.Width].IsBlackHole {
				currentCell.AdjucentBlackHoleQuantity++
			}
			// Down
			if height < field.Height-1 && field.Cells[currentPos+field.Width].IsBlackHole {
				currentCell.AdjucentBlackHoleQuantity++
			}
			// Left-Up
			if width > 0 && height > 0 && field.Cells[currentPos-field.Width-1].IsBlackHole {
				currentCell.AdjucentBlackHoleQuantity++
			}
			// Left-Down
			if width > 0 && height < field.Height-1 && field.Cells[currentPos+field.Width-1].IsBlackHole {
				currentCell.AdjucentBlackHoleQuantity++
			}
			// Right-Up
			if width < field.Width-1 && height > 0 && field.Cells[currentPos-field.Width+1].IsBlackHole {
				currentCell.AdjucentBlackHoleQuantity++
			}
			// Right-Down
			if width < field.Width-1 && height < field.Height-1 && field.Cells[currentPos+field.Width+1].IsBlackHole {
				currentCell.AdjucentBlackHoleQuantity++
			}
		}
	}
}

// printField prints all the cells this way:
//	- '=' if it is exploded
//  - '*' if it is invisible and non-marked
//	- '#' if it is invisible and marked
//  - '@' if it is a black hole
//  - 'n' (n-quantity of adjacting black holes) in the rest of the cases ('0' is substituted with '.')
// (extra space is printed after each symbol because console prints extra space between rows)
func (field *playField) printField() {
	fmt.Fprintln(os.Stdout, "")
	for height := 0; height < field.Height; height++ {
		var builder strings.Builder
		for width := 0; width < field.Width; width++ {
			rowBias := height * field.Width
			currentPos := rowBias + width
			currentCell := &field.Cells[currentPos]
			val := fmt.Sprintf("%d ", currentCell.AdjucentBlackHoleQuantity)
			if currentCell.AdjucentBlackHoleQuantity == 0 {
				val = ". "
			}

			if currentCell.IsExploded {
				val = "= "
			} else if !currentCell.IsVisible {
				if currentCell.IsMarked {
					val = "# "
				} else {
					val = "* "
				}
			} else if currentCell.IsBlackHole {
				val = "@ "
			}
			builder.WriteString(val)
		}
		symdolsLine := builder.String()
		fmt.Fprintln(os.Stdout, symdolsLine)
	}
}

// updateFieldCell updates playField cell in accordance with user's actions
func (field *playField) updateFieldCell(width, height int, marked bool) {
	rowBias := height * field.Width
	currentCell := &field.Cells[rowBias+width]
	if marked {
		currentCell.IsMarked = true
	} else {
		currentCell.IsVisible = true
	}
}

// checkFieldBlackHole checks if a cell black hole should be exploded
func (field *playField) checkFieldBlackHole(width, height int) bool {
	rowBias := height * field.Width
	currentCell := &field.Cells[rowBias+width]

	if currentCell.IsVisible && currentCell.IsBlackHole {
		currentCell.IsExploded = true
		return true
	}
	return false
}

// updateFieldVisibility updates current and the adjacent cells visibility. Particularly, if a cell
// has zero adjacent black holes the game needs to automatically make the surrounding cells visible.
// previousCells is used to remember previous cells and prevent endless recursion.
// Since Go does not have Set collection, use the map of empty structs for previousCells instead.
func (field *playField) updateFieldVisibility(width, height int, previousCells map[playCellPos]struct{}) {
	rowBias := height * field.Width
	currentPos := rowBias + width
	currentCell := &field.Cells[currentPos]

	if currentCell.AdjucentBlackHoleQuantity == 0 {
		// Left
		if width > 0 && !field.Cells[currentPos-1].IsBlackHole {
			field.Cells[currentPos-1].IsVisible = true
			currentPlayCellPos := playCellPos{width: width - 1, height: height}
			if _, ok := previousCells[currentPlayCellPos]; !ok {
				previousCells[currentPlayCellPos] = struct{}{}
				field.updateFieldVisibility(width-1, height, previousCells)
			}
		}
		// Right
		if width < field.Width-1 && !field.Cells[currentPos+1].IsBlackHole {
			field.Cells[currentPos+1].IsVisible = true
			currentPlayCellPos := playCellPos{width: width + 1, height: height}
			if _, ok := previousCells[currentPlayCellPos]; !ok {
				previousCells[currentPlayCellPos] = struct{}{}
				field.updateFieldVisibility(width+1, height, previousCells)
			}
		}
		// Up
		if height > 0 && !field.Cells[currentPos-field.Width].IsBlackHole {
			field.Cells[currentPos-field.Width].IsVisible = true
			currentPlayCellPos := playCellPos{width: width, height: height - 1}
			if _, ok := previousCells[currentPlayCellPos]; !ok {
				previousCells[currentPlayCellPos] = struct{}{}
				field.updateFieldVisibility(width, height-1, previousCells)
			}
		}
		// Down
		if height < field.Height-1 && !field.Cells[currentPos+field.Width].IsBlackHole {
			field.Cells[currentPos+field.Width].IsVisible = true
			currentPlayCellPos := playCellPos{width: width, height: height + 1}
			if _, ok := previousCells[currentPlayCellPos]; !ok {
				previousCells[currentPlayCellPos] = struct{}{}
				field.updateFieldVisibility(width, height+1, previousCells)
			}
		}
	}
}

// checkWinResult check if a player has already won
func (field *playField) checkWinResult() bool {
	for height := 0; height < field.Height; height++ {
		for width := 0; width < field.Width; width++ {
			currentCell := &field.Cells[height*field.Width+width]
			if !currentCell.IsBlackHole && (!currentCell.IsVisible || currentCell.IsMarked) {
				return false
			}
		}
	}
	return true
}

// parseCommandParameters parses user command to reveal to mark a cell
func (field *playField) parseUserCommand(commandParameters string) (int, int, bool, error) {
	trimmedCommandParameters := strings.TrimSpace(commandParameters)
	commandArguments := strings.Fields(trimmedCommandParameters)
	commandArgumentsLen := len(commandArguments)

	if commandArgumentsLen < 2 {
		return 0, 0, false, errors.New("not enough arguments are provided, should be 'width height (marked)'")
	}

	if commandArgumentsLen == 2 || commandArgumentsLen == 3 {
		width, err := strconv.Atoi(commandArguments[0])
		if err != nil {
			return 0, 0, false, fmt.Errorf("failed to convert width(%s) from string to int", commandArguments[0])
		}
		if width >= field.Width {
			return 0, 0, false, fmt.Errorf("entered width(%d) should be less or equal to %d", width, field.Width-1)
		}

		height, err := strconv.Atoi(commandArguments[1])
		if err != nil {
			return 0, 0, false, fmt.Errorf("failed to convert height(%s) from string to int", commandArguments[1])
		}
		if height >= field.Height {
			return 0, 0, false, fmt.Errorf("entered height(%d) should be less or equal to %d", height, field.Height-1)
		}

		if commandArgumentsLen == 3 {
			if commandArguments[2] == "true" || commandArguments[2] == "1" {
				return width, height, true, nil
			}
			if commandArguments[2] == "false" || commandArguments[2] == "0" {
				return width, height, false, nil
			}

			return 0, 0, false, fmt.Errorf("failed to convert marck(%s) from string to bool", commandArguments[2])
		}

		return width, height, false, nil
	}

	return 0, 0, false, errors.New("too many arguments, should be less or equal to 3: 'width height (marked)'")
}
