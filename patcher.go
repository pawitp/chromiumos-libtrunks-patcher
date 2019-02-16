package main

import (
	"bytes"
	"debug/elf"
	"errors"
	"fmt"
	"os"
)

func check(e error) {
	if e != nil {
		fmt.Println("Error: " + e.Error())
		os.Exit(1)
	}
}

func Find(symbols []elf.Symbol, name string) (elf.Symbol, error) {
	for _, s := range symbols {
		if s.Name == name {
			return s, nil
		}
	}
	return elf.Symbol{}, errors.New("Symbol not found " + name)
}

func Patch(data []byte, find []byte, replace []byte) ([]byte, error) {
	if len(find) != len(replace) {
		return nil, errors.New("Size of find does not equal replace")
	}
	index := bytes.Index(data, find)
	if index == -1 {
		return nil, errors.New("Search string not found in data")
	}
	return bytes.Replace(data, find, replace, 1), nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s libtrunks.so\n", os.Args[0])
		os.Exit(1)
	}

	f, err := os.OpenFile(os.Args[1], os.O_RDWR, 0)
	check(err)
	elf, err := elf.NewFile(f)
	check(err)

	symbols, err := elf.DynamicSymbols()
	check(err)

	// Find required symbol
	pheSymbol, err := Find(symbols, "_ZN6trunks12TpmStateImpl26IsPlatformHierarchyEnabledEv")
	check(err)
	fmt.Printf("Found symbol %#v\n", pheSymbol)

	// Read the function to byte
	funcBytes := make([]byte, pheSymbol.Size)
	_, err = f.ReadAt(funcBytes, int64(pheSymbol.Value))
	check(err)

	// Patch the function
	newFuncBytes, err := Patch(funcBytes,
		[]byte{0x75, 0x13, 0x83, 0xE0, 0x01, 0x48, 0x81, 0xC4},
		[]byte{0x75, 0x13, 0x83, 0xE0, 0x00, 0x48, 0x81, 0xC4})
	check(err)

	// Write back to file
	_, err = f.WriteAt(newFuncBytes, int64(pheSymbol.Value))
	check(err)

	err = f.Close()
	check(err)
	fmt.Println("Patched file successfully")
}
