package scorer

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

const int16Scale = float32(32767.0)
const recordSize = dims*2 + 1

func EnsureDatasetBin(jsonPath, binPath, sharedDir string) error {
	lockPath := sharedDir + "/dataset.lock"
	readyPath := sharedDir + "/dataset.ready"

	if _, err := os.Stat(readyPath); err == nil {
		log.Println("dataset.bin já existe, pulando conversão")
		return nil
	}

	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("aguardando conversão do dataset por outra instância...")
		return waitForReady(readyPath)
	}
	lockFile.Close()

	log.Println("convertendo references.json → dataset.bin...")
	if err := convertToBin(jsonPath, binPath); err != nil {
		os.Remove(lockPath)
		return fmt.Errorf("erro na conversão: %w", err)
	}

	ready, err := os.Create(readyPath)
	if err != nil {
		return err
	}
	ready.Close()
	os.Remove(lockPath)

	log.Println("conversão concluída")
	return nil
}

func waitForReady(readyPath string) error {
	timeout := time.After(10 * time.Minute)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout esperando dataset.bin")
		case <-ticker.C:
			if _, err := os.Stat(readyPath); err == nil {
				log.Println("dataset.bin pronto")
				return nil
			}
		}
	}
}

func convertToBin(jsonPath, binPath string) error {
	src, err := os.Open(jsonPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(binPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	decoder := json.NewDecoder(src)
	if _, err := decoder.Token(); err != nil {
		return err
	}

	count := 0
	buf := make([]byte, recordSize)

	for decoder.More() {
		var ref Reference
		if err := decoder.Decode(&ref); err != nil {
			return err
		}

		for d := 0; d < dims; d++ {
			val := int16(ref.Vector[d] * int16Scale)
			binary.LittleEndian.PutUint16(buf[d*2:], uint16(val))
		}

		if ref.Label == "fraud" {
			buf[dims*2] = 1
		} else {
			buf[dims*2] = 0
		}

		if _, err := dst.Write(buf); err != nil {
			return err
		}
		count++
	}

	log.Printf("%d vetores convertidos\n", count)
	return nil
}
