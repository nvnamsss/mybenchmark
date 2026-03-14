package mybenchmark

import (
	"io"
	"os"
	"testing"
)

// BenchmarkDiskWriteSmall benchmarks small file write (5 bytes)
func BenchmarkDiskWriteSmall(b *testing.B) {
	data := []byte("hello")
	tmpDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file, err := os.Create(tmpDir + "/test_small.dat")
		if err != nil {
			b.Fatal(err)
		}
		_, err = file.Write(data)
		if err != nil {
			b.Fatal(err)
		}
		file.Close()
	}
}

// BenchmarkDiskWriteMedium benchmarks medium file write (4KB)
func BenchmarkDiskWriteMedium(b *testing.B) {
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i % 256)
	}
	tmpDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file, err := os.Create(tmpDir + "/test_medium.dat")
		if err != nil {
			b.Fatal(err)
		}
		_, err = file.Write(data)
		if err != nil {
			b.Fatal(err)
		}
		file.Close()
	}
}

// BenchmarkDiskWriteLarge benchmarks large file write (1MB)
func BenchmarkDiskWriteLarge(b *testing.B) {
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	tmpDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file, err := os.Create(tmpDir + "/test_large.dat")
		if err != nil {
			b.Fatal(err)
		}
		_, err = file.Write(data)
		if err != nil {
			b.Fatal(err)
		}
		file.Close()
	}
}

// BenchmarkDiskReadSmall benchmarks small file read (5 bytes)
func BenchmarkDiskReadSmall(b *testing.B) {
	data := []byte("hello")
	tmpDir := b.TempDir()
	testFile := tmpDir + "/test_read_small.dat"

	// Create test file
	file, err := os.Create(testFile)
	if err != nil {
		b.Fatal(err)
	}
	file.Write(data)
	file.Close()

	buf := make([]byte, len(data))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file, err := os.Open(testFile)
		if err != nil {
			b.Fatal(err)
		}
		_, err = file.Read(buf)
		if err != nil {
			b.Fatal(err)
		}
		file.Close()
	}
}

// BenchmarkDiskReadMedium benchmarks medium file read (4KB)
func BenchmarkDiskReadMedium(b *testing.B) {
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i % 256)
	}
	tmpDir := b.TempDir()
	testFile := tmpDir + "/test_read_medium.dat"

	// Create test file
	file, err := os.Create(testFile)
	if err != nil {
		b.Fatal(err)
	}
	file.Write(data)
	file.Close()

	buf := make([]byte, len(data))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file, err := os.Open(testFile)
		if err != nil {
			b.Fatal(err)
		}
		_, err = io.ReadFull(file, buf)
		if err != nil {
			b.Fatal(err)
		}
		file.Close()
	}
}

// BenchmarkDiskReadLarge benchmarks large file read (1MB)
func BenchmarkDiskReadLarge(b *testing.B) {
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	tmpDir := b.TempDir()
	testFile := tmpDir + "/test_read_large.dat"

	// Create test file
	file, err := os.Create(testFile)
	if err != nil {
		b.Fatal(err)
	}
	file.Write(data)
	file.Close()

	buf := make([]byte, len(data))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file, err := os.Open(testFile)
		if err != nil {
			b.Fatal(err)
		}
		_, err = io.ReadFull(file, buf)
		if err != nil {
			b.Fatal(err)
		}
		file.Close()
	}
}

// BenchmarkDiskWriteSync benchmarks write with sync (force to disk)
func BenchmarkDiskWriteSync(b *testing.B) {
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i % 256)
	}
	tmpDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file, err := os.Create(tmpDir + "/test_sync.dat")
		if err != nil {
			b.Fatal(err)
		}
		_, err = file.Write(data)
		if err != nil {
			b.Fatal(err)
		}
		file.Sync() // Force write to disk
		file.Close()
	}
}

// BenchmarkDiskSequentialWrite benchmarks sequential writes to same file
func BenchmarkDiskSequentialWrite(b *testing.B) {
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i % 256)
	}
	tmpDir := b.TempDir()
	testFile := tmpDir + "/test_sequential.dat"

	file, err := os.Create(testFile)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = file.Write(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDiskSequentialRead benchmarks sequential reads from same file
func BenchmarkDiskSequentialRead(b *testing.B) {
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i % 256)
	}
	tmpDir := b.TempDir()
	testFile := tmpDir + "/test_seq_read.dat"

	// Create large test file
	file, err := os.Create(testFile)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < 10000; i++ {
		file.Write(data)
	}
	file.Close()

	file, err = os.Open(testFile)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	buf := make([]byte, 4096)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = io.ReadFull(file, buf)
		if err != nil {
			if err == io.EOF {
				file.Seek(0, 0) // Reset to beginning
				continue
			}
			b.Fatal(err)
		}
	}
}

// BenchmarkDiskRandomRead benchmarks random reads (seeking)
func BenchmarkDiskRandomRead(b *testing.B) {
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i % 256)
	}
	tmpDir := b.TempDir()
	testFile := tmpDir + "/test_random_read.dat"

	// Create test file with multiple blocks
	file, err := os.Create(testFile)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < 1000; i++ {
		file.Write(data)
	}
	file.Close()

	file, err = os.Open(testFile)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	buf := make([]byte, 4096)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := int64((i % 1000) * 4096)
		file.Seek(offset, 0)
		_, err = io.ReadFull(file, buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}
