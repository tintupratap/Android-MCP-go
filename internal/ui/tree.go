package ui

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var InteractiveClasses = map[string]bool{
	"android.widget.Button":      true,
	"android.widget.ImageButton": true,
	"android.widget.EditText":    true,
	"android.widget.CheckBox":    true,
	"android.widget.Switch":      true,
	"android.widget.RadioButton": true,
	"android.widget.Spinner":     true,
	"android.widget.SeekBar":     true,
}

type BoundingBox struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
}

func (b BoundingBox) ToString() string {
	return fmt.Sprintf("[%d,%d][%d,%d]", b.X1, b.Y1, b.X2, b.Y2)
}

type CenterCord struct {
	X int
	Y int
}

func (c CenterCord) ToString() string {
	return fmt.Sprintf("(%d,%d)", c.X, c.Y)
}

type ElementNode struct {
	Name        string
	ClassName   string
	Coordinates CenterCord
	BoundingBox BoundingBox
	ResourceID  string
}

type TreeState struct {
	InteractiveElements []ElementNode
}

func (ts TreeState) ToString() string {
	headers := []string{"Label", "Name", "ResourceId", "Class", "Coordinates"}
	var rows [][]string

	for i, node := range ts.InteractiveElements {
		rows = append(rows, []string{
			strconv.Itoa(i),
			node.Name,
			node.ResourceID,
			node.ClassName,
			node.Coordinates.ToString(),
		})
	}
	return FormatTable(headers, rows)
}

type XMLNode struct {
	XMLName       xml.Name  `xml:"node"`
	Index         string    `xml:"index,attr"`
	Text          string    `xml:"text,attr"`
	ResourceID    string    `xml:"resource-id,attr"`
	Class         string    `xml:"class,attr"`
	Package       string    `xml:"package,attr"`
	ContentDesc   string    `xml:"content-desc,attr"`
	Checkable     string    `xml:"checkable,attr"`
	Checked       string    `xml:"checked,attr"`
	Clickable     string    `xml:"clickable,attr"`
	Enabled       string    `xml:"enabled,attr"`
	Focusable     string    `xml:"focusable,attr"`
	Focused       string    `xml:"focused,attr"`
	Scrollable    string    `xml:"scrollable,attr"`
	LongClickable string    `xml:"long-clickable,attr"`
	Password      string    `xml:"password,attr"`
	Selected      string    `xml:"selected,attr"`
	Bounds        string    `xml:"bounds,attr"`
	Children      []XMLNode `xml:"node"`
}

var boundsRegex = regexp.MustCompile(`\[(\d+),(\d+)\]\[(\d+),(\d+)\]`)

func ExtractCoordinates(bounds string) (int, int, int, int, bool) {
	matches := boundsRegex.FindStringSubmatch(bounds)
	if len(matches) < 5 {
		return 0, 0, 0, 0, false
	}
	x1, _ := strconv.Atoi(matches[1])
	y1, _ := strconv.Atoi(matches[2])
	x2, _ := strconv.Atoi(matches[3])
	y2, _ := strconv.Atoi(matches[4])
	return x1, y1, x2, y2, true
}

func GetCenterCoordinates(x1, y1, x2, y2 int) (int, int) {
	return (x1 + x2) / 2, (y1 + y2) / 2
}

func IsInteractive(node *XMLNode) bool {
	return node.Focusable == "true" ||
		node.Clickable == "true" ||
		node.LongClickable == "true" ||
		node.Checkable == "true" ||
		node.Scrollable == "true" ||
		node.Selected == "true" ||
		node.Password == "true" ||
		InteractiveClasses[node.Class]
}

func GetElementName(node *XMLNode) string {
	name := strings.TrimSpace(node.ContentDesc)
	if name == "" {
		name = strings.TrimSpace(node.Text)
	}
	if name != "" {
		return name
	}

	var primaryTexts []string
	var fallbackTexts []string

	var collectText func(n *XMLNode, isRoot bool)
	collectText = func(n *XMLNode, isRoot bool) {
		isActionable := !isRoot && (n.Clickable == "true" || n.LongClickable == "true" || n.Checkable == "true" || n.Scrollable == "true")

		val := strings.TrimSpace(n.Text)
		if val == "" {
			val = strings.TrimSpace(n.ContentDesc)
		}

		if isActionable {
			if val != "" {
				fallbackTexts = append(fallbackTexts, val)
			}
			return
		}

		if val != "" {
			primaryTexts = append(primaryTexts, val)
		}

		for i := range n.Children {
			collectText(&n.Children[i], false)
		}
	}

	collectText(node, true)

	finalTexts := primaryTexts
	if len(finalTexts) == 0 {
		finalTexts = fallbackTexts
	}
	return strings.TrimSpace(strings.Join(finalTexts, " "))
}

type Hierarchy struct {
	XMLName  xml.Name  `xml:"hierarchy"`
	Rotation string    `xml:"rotation,attr"`
	Nodes    []XMLNode `xml:"node"`
}

func unmarshalHierarchy(xmlData string) ([]XMLNode, error) {
	var h Hierarchy
	if err := xml.Unmarshal([]byte(xmlData), &h); err == nil && len(h.Nodes) > 0 {
		return h.Nodes, nil
	}

	var root XMLNode
	if err := xml.Unmarshal([]byte(xmlData), &root); err == nil {
		return []XMLNode{root}, nil
	}

	return nil, fmt.Errorf("failed to parse xml hierarchy")
}

func ParseTreeState(xmlData string) (*TreeState, error) {
	nodes, err := unmarshalHierarchy(xmlData)
	if err != nil {
		return nil, err
	}

	var interactive []ElementNode

	var traverse func(node *XMLNode)
	traverse = func(node *XMLNode) {
		if node.Enabled == "true" && IsInteractive(node) {
			x1, y1, x2, y2, ok := ExtractCoordinates(node.Bounds)
			if ok {
				name := GetElementName(node)
				if name != "" {
					cx, cy := GetCenterCoordinates(x1, y1, x2, y2)
					rawID := node.ResourceID
					shortID := rawID
					if idx := strings.Index(rawID, "/"); idx != -1 {
						shortID = rawID[idx+1:]
					}

					interactive = append(interactive, ElementNode{
						Name:        name,
						ClassName:   node.Class,
						Coordinates: CenterCord{X: cx, Y: cy},
						BoundingBox: BoundingBox{X1: x1, Y1: y1, X2: x2, Y2: y2},
						ResourceID:  shortID,
					})
				}
			}
		}

		for i := range node.Children {
			traverse(&node.Children[i])
		}
	}

	for i := range nodes {
		traverse(&nodes[i])
	}
	return &TreeState{InteractiveElements: interactive}, nil
}

type Selector struct {
	Text        string
	ResourceID  string
	ClassName   string
	Description string
	Index       int
}

func MatchSelector(node *XMLNode, sel Selector, pkg string) bool {
	if sel.Text != "" && node.Text != sel.Text {
		return false
	}
	if sel.ClassName != "" && node.Class != sel.ClassName {
		return false
	}
	if sel.Description != "" && node.ContentDesc != sel.Description {
		return false
	}
	if sel.ResourceID != "" {
		resID := sel.ResourceID
		if pkg != "" && !strings.Contains(resID, "/") && !strings.Contains(resID, ":") {
			resID = fmt.Sprintf("%s:id/%s", pkg, resID)
		}
		if node.ResourceID != resID && node.ResourceID != sel.ResourceID {
			// Check if short ID matches
			shortID := node.ResourceID
			if idx := strings.Index(shortID, "/"); idx != -1 {
				shortID = shortID[idx+1:]
			}
			if shortID != sel.ResourceID {
				return false
			}
		}
	}
	return true
}

func FindElementBySelector(xmlData string, sel Selector, pkg string) (*ElementNode, error) {
	nodes, err := unmarshalHierarchy(xmlData)
	if err != nil {
		return nil, err
	}

	var matches []XMLNode
	var traverse func(n *XMLNode)
	traverse = func(n *XMLNode) {
		if MatchSelector(n, sel, pkg) {
			matches = append(matches, *n)
		}
		for i := range n.Children {
			traverse(&n.Children[i])
		}
	}

	for i := range nodes {
		traverse(&nodes[i])
	}
	if len(matches) == 0 {
		return nil, nil
	}

	targetIndex := sel.Index
	if targetIndex < 0 || targetIndex >= len(matches) {
		targetIndex = 0
	}

	target := matches[targetIndex]
	x1, y1, x2, y2, ok := ExtractCoordinates(target.Bounds)
	if !ok {
		return nil, fmt.Errorf("element found but bounds invalid: %s", target.Bounds)
	}
	cx, cy := GetCenterCoordinates(x1, y1, x2, y2)
	name := GetElementName(&target)
	shortID := target.ResourceID
	if idx := strings.Index(shortID, "/"); idx != -1 {
		shortID = shortID[idx+1:]
	}

	return &ElementNode{
		Name:        name,
		ClassName:   target.Class,
		Coordinates: CenterCord{X: cx, Y: cy},
		BoundingBox: BoundingBox{X1: x1, Y1: y1, X2: x2, Y2: y2},
		ResourceID:  shortID,
	}, nil
}
