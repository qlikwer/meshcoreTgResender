package main

import (
	"fmt"
	"strings"
)

// routeNode — узел дерева маршрутов. Дети хранятся в порядке первого
// появления (важно для итогового вида дерева: внутри уровня сначала идут
// все конечные узлы (наблюдатели) в порядке обнаружения, затем — ветки,
// тоже в порядке обнаружения).
type routeNode struct {
	label    string
	children []*routeNode
	index    map[string]int
}

func newRouteNode(label string) *routeNode {
	return &routeNode{label: label, index: make(map[string]int)}
}

func (n *routeNode) getOrAdd(label string) *routeNode {
	if i, ok := n.index[label]; ok {
		return n.children[i]
	}
	c := newRouteNode(label)
	n.index[label] = len(n.children)
	n.children = append(n.children, c)
	return c
}

// buildRouteTree строит дерево из набора цепочек (например
// [][]string{{"47","31","80","3A","8077"}, ...}), полученных из
// display_combined_path. Порядок chains должен соответствовать порядку
// первого появления маршрута (хронологический), иначе сортировка
// "листья сначала, ветки потом" внутри уровня может оказаться случайной.
func buildRouteTree(chains [][]string) *routeNode {
	root := newRouteNode("")
	for _, chain := range chains {
		cur := root
		for _, seg := range chain {
			cur = cur.getOrAdd(seg)
		}
	}
	return root
}

// orderedChildren: сначала листья (узлы без потомков — конечные наблюдатели),
// потом ветки — в каждой группе сохраняется порядок первого появления.
func orderedChildren(n *routeNode) []*routeNode {
	ordered := make([]*routeNode, 0, len(n.children))
	var branches []*routeNode
	for _, c := range n.children {
		if len(c.children) == 0 {
			ordered = append(ordered, c)
		} else {
			branches = append(branches, c)
		}
	}
	return append(ordered, branches...)
}

// renderRouteNode рекурсивно рисует поддерево в виде ASCII-арта (как
// `tree`, но горизонтально). Возвращает готовые строки и индекс строки,
// на которой стоит подпись текущего узла — это нужно родителю, чтобы
// нарисовать свой коннектор на той же высоте.
func renderRouteNode(n *routeNode, labelOverride string) ([]string, int) {
	children := orderedChildren(n)

	if len(children) == 0 {
		return []string{fmt.Sprintf("[%s]", n.label)}, 0
	}

	type block struct {
		lines  []string
		anchor int
		offset int
	}

	blocks := make([]block, len(children))
	var combined []string
	for i, c := range children {
		lines, anchor := renderRouteNode(c, "")
		blocks[i] = block{lines: lines, anchor: anchor, offset: len(combined)}
		combined = append(combined, lines...)
		if i != len(children)-1 {
			combined = append(combined, "")
		}
	}

	count := len(children)
	anchorRows := make([]int, count)
	for i, b := range blocks {
		anchorRows[i] = b.offset + b.anchor
	}

	lo, hi := anchorRows[0], anchorRows[0]
	for _, r := range anchorRows {
		if r < lo {
			lo = r
		}
		if r > hi {
			hi = r
		}
	}

	out := make([]string, len(combined))
	for r, content := range combined {
		var prefix string
		switch idx := indexOfInt(anchorRows, r); {
		case idx >= 0 && count == 1:
			prefix = "── "
		case idx == 0:
			prefix = "┌─ "
		case idx == count-1:
			prefix = "└─ "
		case idx > 0:
			prefix = "├─ "
		case r > lo && r < hi:
			prefix = "│  "
		default:
			prefix = "   "
		}
		out[r] = prefix + content
	}

	label := n.label
	if labelOverride != "" {
		label = labelOverride
	}
	const suffix = "─"
	width := len([]rune(label)) + 1 + len([]rune(suffix))

	anchorRow := anchorRows[count/2]

	final := make([]string, len(out))
	for r, line := range out {
		if r == anchorRow {
			final[r] = fmt.Sprintf("%s %s%s", label, suffix, line)
		} else {
			final[r] = strings.Repeat(" ", width) + line
		}
	}
	return final, anchorRow
}

func indexOfInt(s []int, v int) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return -1
}

// parseRouteChains разбирает список значений display_combined_path вида
// "NA → 47 → 31 → 80 → 3A → 8077" в цепочки сегментов без "NA":
// []string{"47","31","80","3A","8077"}. Порядок paths должен совпадать с
// порядком первого появления маршрута.
func parseRouteChains(paths []string) [][]string {
	chains := make([][]string, 0, len(paths))
	for _, p := range paths {
		parts := strings.Split(p, "→")
		seg := make([]string, 0, len(parts))
		for _, s := range parts {
			s = strings.TrimSpace(s)
			if s == "" || s == "NA" {
				continue
			}
			seg = append(seg, s)
		}
		if len(seg) > 0 {
			chains = append(chains, seg)
		}
	}
	return chains
}

// RenderRouteTree строит и отрисовывает ASCII-дерево маршрутов одного
// сообщения: rootLabel — имя отправителя (корень дерева), paths — все
// display_combined_path, по которым это сообщение долетело до наблюдателей,
// в порядке первого появления.
func RenderRouteTree(rootLabel string, paths []string) string {
	chains := parseRouteChains(paths)
	if len(chains) == 0 {
		return fmt.Sprintf("[%s]", rootLabel)
	}
	root := buildRouteTree(chains)
	lines, _ := renderRouteNode(root, rootLabel)
	return strings.Join(lines, "\n")
}
