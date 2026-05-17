//go:build !windows

package scorer

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"sync"
	"syscall"
)

const dims = 14
const numPlanes = 16

type Reference struct {
	Vector []float32 `json:"vector"`
	Label  string    `json:"label"`
}

type SearchResult struct {
	Similarity float32
	Label      string
}

type HNSWIndex struct {
	data    []byte             // arquivo mapeado via mmap
	buckets map[uint32][]int32 // LSH: assinatura → posições no arquivo
	planes  [][]float32        // hiperplanos LSH
	numCPU  int
	mu      sync.RWMutex
}

func LoadMmapIndex(binPath string) (*HNSWIndex, error) {
	file, err := os.Open(binPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	size := info.Size()
	if size%int64(recordSize) != 0 {
		return nil, fmt.Errorf("arquivo corrompido: tamanho inválido")
	}

	count := int(size / int64(recordSize))
	log.Printf("mapeando %d vetores via mmap\n", count)

	// mapeia o arquivo — compartilhado entre processos via page cache
	data, err := syscall.Mmap(
		int(file.Fd()),
		0,
		int(size),
		syscall.PROT_READ,
		syscall.MAP_SHARED,
	)
	if err != nil {
		return nil, fmt.Errorf("erro no mmap: %w", err)
	}

	// gera hiperplanos LSH
	planes := make([][]float32, numPlanes)
	for i := range planes {
		planes[i] = randomPlane(dims)
	}

	numCPU := runtime.NumCPU()
	if numCPU > 2 {
		numCPU = 2
	}

	index := &HNSWIndex{
		data:    data,
		buckets: make(map[uint32][]int32),
		planes:  planes,
		numCPU:  numCPU,
	}

	// constrói o índice LSH — guarda só posições, não os vetores
	log.Println("construindo índice LSH...")
	for i := 0; i < count; i++ {
		vec := index.getVector(i)
		sig := index.signature(vec)
		index.buckets[sig] = append(index.buckets[sig], int32(i))
	}

	log.Printf("índice LSH pronto — %d baldes\n", len(index.buckets))
	return index, nil
}

// getVector lê um vetor do mmap e converte int16 → float32
func (h *HNSWIndex) getVector(i int) []float32 {
	offset := i * recordSize
	vec := make([]float32, dims)
	for d := 0; d < dims; d++ {
		raw := int16(binary.LittleEndian.Uint16(h.data[offset+d*2:]))
		vec[d] = float32(raw) / int16Scale
	}
	return vec
}

func (h *HNSWIndex) getLabel(i int) string {
	offset := i*recordSize + dims*2
	if h.data[offset] == 1 {
		return "fraud"
	}
	return "legit"
}

func (h *HNSWIndex) signature(vec []float32) uint32 {
	var sig uint32
	for i, plane := range h.planes {
		if dotProduct(vec, plane) >= 0 {
			sig |= 1 << uint(i)
		}
	}
	return sig
}

func (h *HNSWIndex) Search(query []float32, k int) []SearchResult {
	sig := h.signature(query)

	h.mu.RLock()
	positions, ok := h.buckets[sig]
	h.mu.RUnlock()

	// expande para baldes vizinhos se balde vazio ou pequeno
	if !ok || len(positions) < k {
		positions = h.expandSearch(sig, k*20)
	}

	if len(positions) == 0 {
		return []SearchResult{}
	}

	return h.parallelSearch(query, positions, k)
}

func (h *HNSWIndex) expandSearch(sig uint32, max int) []int32 {
	positions := []int32{}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if p, ok := h.buckets[sig]; ok {
		positions = append(positions, p...)
	}

	for bit := 0; bit < numPlanes && len(positions) < max; bit++ {
		neighborSig := sig ^ (1 << uint(bit))
		if p, ok := h.buckets[neighborSig]; ok {
			positions = append(positions, p...)
		}
	}

	return positions
}

func (h *HNSWIndex) parallelSearch(query []float32, positions []int32, k int) []SearchResult {
	total := len(positions)
	numWorkers := h.numCPU
	if numWorkers > total {
		numWorkers = total
	}

	chunkSize := total / numWorkers
	results := make([]SearchResult, total)
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		start := w * chunkSize
		end := start + chunkSize
		if w == numWorkers-1 {
			end = total
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				pos := int(positions[i])
				vec := h.getVector(pos)
				results[i] = SearchResult{
					Similarity: cosineSimilarity(query, vec),
					Label:      h.getLabel(pos),
				}
			}
		}(start, end)
	}

	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	if k > len(results) {
		k = len(results)
	}
	return results[:k]
}

func randomPlane(d int) []float32 {
	plane := make([]float32, d)
	for i := range plane {
		plane[i] = float32(rand.NormFloat64())
	}
	return plane
}

func dotProduct(a, b []float32) float32 {
	var sum float32
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

func cosineSimilarity(a, b []float32) float32 {
	var dot, normA, normB float32
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return float32(dot / float32(math.Sqrt(float64(normA))*math.Sqrt(float64(normB))))
}
