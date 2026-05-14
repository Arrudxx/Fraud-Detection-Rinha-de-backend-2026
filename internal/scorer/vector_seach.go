package scorer

import (
	"math"
	"runtime"
	"sort"
)

type Reference struct {
	Vector []float64 `json:"vector"`
	Label  string    `json:"label"`
}

type SearchResult struct {
	Similarity float64
	Label      string
}

// ── KD-Tree ──────────────────────────────────────────

type kdNode struct {
	ref   Reference
	left  *kdNode
	right *kdNode
	axis  int // qual dimensão foi usada pra dividir
}

type HNSWIndex struct {
	root   *kdNode
	numCPU int
}

func BuildIndex(refs []Reference) (*HNSWIndex, error) {
	nodes := make([]Reference, len(refs))
	copy(nodes, refs)

	root := buildKDTree(nodes, 0)

	return &HNSWIndex{
		root:   root,
		numCPU: runtime.NumCPU(),
	}, nil
}

// buildKDTree constrói a árvore recursivamente
func buildKDTree(refs []Reference, depth int) *kdNode {
	if len(refs) == 0 {
		return nil
	}

	// escolhe qual dimensão dividir (alterna a cada nível)
	axis := depth % len(refs[0].Vector)

	// ordena pelos valores dessa dimensão
	sort.Slice(refs, func(i, j int) bool {
		return refs[i].Vector[axis] < refs[j].Vector[axis]
	})

	// pega o ponto do meio como raiz desse nível
	mid := len(refs) / 2

	return &kdNode{
		ref:   refs[mid],
		axis:  axis,
		left:  buildKDTree(refs[:mid], depth+1),   // metade menor
		right: buildKDTree(refs[mid+1:], depth+1), // metade maior
	}
}

// ── Busca ─────────────────────────────────────────────

func (h *HNSWIndex) Search(query []float64, k int) []SearchResult {
	// 1. KD-Tree encontra candidatos na região certa
	candidates := h.searchKDTree(query, k*20) // pega 20x K pra ter margem

	// 2. goroutines refinam entre os candidatos
	return h.parallelSearch(query, candidates, k)
}

func (h *HNSWIndex) searchKDTree(query []float64, maxCandidates int) []Reference {
	candidates := make([]Reference, 0, maxCandidates)
	h.traverse(h.root, query, &candidates, maxCandidates)
	return candidates
}

// traverse desce pela árvore coletando candidatos
func (h *HNSWIndex) traverse(node *kdNode, query []float64, candidates *[]Reference, max int) {
	if node == nil {
		return
	}

	*candidates = append(*candidates, node.ref)

	// decide qual lado da árvore é mais promissor
	goLeft := query[node.axis] <= node.ref.Vector[node.axis]

	if goLeft {
		h.traverse(node.left, query, candidates, max)
	} else {
		h.traverse(node.right, query, candidates, max)
	}

	// se ainda não tem candidatos suficientes, explora o outro lado também
	if len(*candidates) < max {
		if goLeft {
			h.traverse(node.right, query, candidates, max)
		} else {
			h.traverse(node.left, query, candidates, max)
		}
	}
}

// parallelSearch usa goroutines para calcular similaridade nos candidatos
func (h *HNSWIndex) parallelSearch(query []float64, candidates []Reference, k int) []SearchResult {
	total := len(candidates)
	numWorkers := h.numCPU
	if numWorkers > total {
		numWorkers = total
	}

	chunkSize := total / numWorkers
	results := make([]SearchResult, total)
	done := make(chan struct{}, numWorkers)

	for w := 0; w < numWorkers; w++ {
		start := w * chunkSize
		end := start + chunkSize
		if w == numWorkers-1 {
			end = total
		}

		go func(start, end int) {
			for i := start; i < end; i++ {
				results[i] = SearchResult{
					Similarity: cosineSimilarity(query, candidates[i].Vector),
					Label:      candidates[i].Label,
				}
			}
			done <- struct{}{}
		}(start, end)
	}

	for w := 0; w < numWorkers; w++ {
		<-done
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	if k > len(results) {
		k = len(results)
	}
	return results[:k]
}

func cosineSimilarity(a, b []float64) float64 {
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func toFloat32(in []float64) []float32 {
	out := make([]float32, len(in))
	for i, v := range in {
		out[i] = float32(v)
	}
	return out
}
