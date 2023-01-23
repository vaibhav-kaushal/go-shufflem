package main

import (
	"errors"
	"fmt"
	"log"
)

const emptyByte byte = 0b00000000

// Masks to set a particular bit to 1
const (
	mask0 byte = 0b10000000
	mask1 byte = 0b01000000
	mask2 byte = 0b00100000
	mask3 byte = 0b00010000
	mask4 byte = 0b00001000
	mask5 byte = 0b00000100
	mask6 byte = 0b00000010
	mask7 byte = 0b00000001
)

// Masks to set a particular bit to 0
const (
	clearMask0 byte = 0b01111111
	clearMask1 byte = 0b10111111
	clearMask2 byte = 0b11011111
	clearMask3 byte = 0b11101111
	clearMask4 byte = 0b11110111
	clearMask5 byte = 0b11111011
	clearMask6 byte = 0b11111101
	clearMask7 byte = 0b11111110
)

// Number of bits we can process in this program
const (
	minBitsToProcess uint16 = 8
	maxBitsToProcess uint16 = 1024
)

type Config struct {
	BitCount   uint16 // Support up to 32767 bits (32 KiB) max
	ShuffleMap map[uint16]uint16
}

// WorkUnit is a single work unit of shuffle
type WorkUnit struct {
	Input        []byte          // The input byte slice
	BitSetterMap map[uint16]bool // The map bit position with their values (boolean) that should go in final result
	Output       []byte          // The output byte slice will be put here
	Validated    bool            // Is the shuffle map and input valid according to config
	Config       Config          // The config, as received by the program.
}

// Shuffle is the main function which will shuffle the bits
func (w *WorkUnit) Shuffle() error {
	err := w.validateConfig()
	if err != nil {
		return fmt.Errorf("E#B0FY3 - Config not validated! Error: %v", err)
	}
	err = w.buildBitSetterMap()
	if err != nil {
		return fmt.Errorf("E#18BQUA - Could not build the setter map. Error: %v", err)
	}
	var result []byte
	result = w.Input
	for i := uint16(0); i < w.Config.BitCount; i++ {
		if bitToSetBool, ok := w.BitSetterMap[i]; ok == true {
			// Found the bit to set in the setter map
			result, err = setBitOnByteArray(bitToSetBool, result, i)
			if err != nil {
				return errors.New(fmt.Sprintf("E#ATZUS - Could not set bit on byte array: %v", err))
			}
		}
	}

	w.Output = result

	return nil
}

// validateConfig validates the config in the WorkUnit
func (w *WorkUnit) validateConfig() error {
	// Ensure that the number of bytes in the Config are not 0 or greater than 128
	if w.Config.BitCount < minBitsToProcess {
		w.Validated = false
		return errors.New(fmt.Sprintf("E#9R6OP - need at least %v bits for the swapping to be done", minBitsToProcess))
	}

	if w.Config.BitCount > maxBitsToProcess {
		w.Validated = false
		return errors.New(fmt.Sprintf("E#9R6Q8 - cannot process more than %v bits for now", maxBitsToProcess))
	}

	// Ensure that the number of bits mentioned is exactly divisible by 8
	if w.Config.BitCount%8 != 0 {
		w.Validated = false
		return errors.New(fmt.Sprintf("E#9R6YV - number of bits to process must be a multiple of 8. Got: %v", w.Config.BitCount))
	}

	// Ensure that the input is exactly the size mentioned in the Config
	if len(w.Input)*8 != int(w.Config.BitCount) {
		w.Validated = false
		return errors.New(fmt.Sprintf("E#9R71S - expected %v bits in input, got %v", w.Config.BitCount, len(w.Input)*8))
	}

	// Ensure that the list of Shufflings that we have to do have no logical overlapping
	var all []uint16
	var exists = false
	var repeatedBitPosition uint16
	// The way this works is: for each shuffle-map entry, we shove both indexes into a single array of
	// index positions (the `all` array). If another bit position is found in any future map entry,
	// it would already exist in  `all` array and will be detected by the element existence search
	for key, value := range w.Config.ShuffleMap {
		if elementExistsInSlice(key, all) {
			exists = true
			repeatedBitPosition = key
			break
		} else {
			all = append(all, key)
		}

		if elementExistsInSlice(value, all) {
			exists = true
			repeatedBitPosition = value
			break
		} else {
			all = append(all, value)
		}
	}

	if exists {
		w.Validated = false
		return errors.New(fmt.Sprintf("E#B0G5P - Cannot continue -- Repetitions in the shuffling map for bit position: %v", repeatedBitPosition))
	}

	w.Validated = true
	return nil
}

// buildBitSetterMap builds the BitSetterMap in the WorkUnit
func (w *WorkUnit) buildBitSetterMap() error {
	for key, value := range w.Config.ShuffleMap {
		err := errors.New("E#B0G0N")

		valAtValueIndex, err := getBit(w.Input[int(value/8)], value%8)
		if err != nil {
			return fmt.Errorf("E#AT666 - Error from getBit: %v", err)
		}
		valAtKeyIndex, err := getBit(w.Input[key/8], key%8)
		if err != nil {
			return fmt.Errorf("E#AT6C0 - Error from getBit: %v", err)
		}

		if valAtKeyIndex != valAtValueIndex {
			w.BitSetterMap[key] = valAtValueIndex
			w.BitSetterMap[value] = valAtKeyIndex
		}
	}
	return nil
}

// buildBitSetterMap builds the BitSetterMap in the WorkUnit
func (w *WorkUnit) buildFullBitMap() {
	for key, value := range w.Config.ShuffleMap {
		err := errors.New("E#O1HYH")
		val, err := getBit(w.Input[int(value/8)], value%8)
		if err != nil {
			fmt.Println("E#AT666 - Error from getBit:", err)
		} else {
			w.BitSetterMap[key] = val
		}

		val, err = getBit(w.Input[key/8], key%8)
		if err != nil {
			fmt.Println("E#AT6C0 - Error from getBit:", err)
		} else {
			w.BitSetterMap[value] = val
		}
	}
	for i := uint16(0); i < w.Config.BitCount; i++ {
		_, ok := w.BitSetterMap[i]
		if !ok {
			// The value was not in the map. So get it from the original input
			ter, err := getBitFromByteArray(w.Input, i)
			if err == nil {
				w.BitSetterMap[i] = ter
			}
		}
	}
}

// elementExistsInSlice tells if a given element exists in a slice
func elementExistsInSlice(element uint16, ins []uint16) bool {
	for _, i := range ins {
		if element == i {
			return true
		}
	}
	return false
}

// getBitFromByteArray gets one bit from the onByteSlice at the give atPosition
func getBitFromByteArray(onByteSlice []byte, atPosition uint16) (bool, error) {
	// Check that the atPosition is valid
	byteSliceLength := len(onByteSlice)
	if atPosition < 0 || atPosition > uint16(byteSliceLength*8)-1 {
		return false, errors.New(fmt.Sprintf("E#AU1SC - At position not within acceptable range: %v", atPosition))
	}

	// Calculate byte's index in the slice
	byteIndexInSlice := atPosition / 8
	bitIndexInByte := atPosition % 8

	// Extract byte
	byteToGetBitFrom := onByteSlice[byteIndexInSlice]

	bitToReturn, err := getBit(byteToGetBitFrom, bitIndexInByte)
	if err != nil {
		return false, errors.New(fmt.Sprintf("E#AU07G - Could not get bit from byte: %v", err))
	}

	return bitToReturn, nil
}

// setBitOnByteArray sets the bit at the atPosition in the onByteSlice and return the resulting byte slice.
func setBitOnByteArray(bit bool, onByteSlice []byte, atPosition uint16) ([]byte, error) {
	// Check that the atPosition is valid
	byteSliceLength := len(onByteSlice)
	if atPosition < 0 || atPosition > uint16(byteSliceLength*8)-1 {
		return onByteSlice, errors.New("E#ATYR9 - At position is not within acceptable range")
	}

	// Calculate byte's index in the slice
	byteIndexInSlice := atPosition / 8
	bitIndexInByte := atPosition % 8

	// Extract byte
	byteToSetBitOn := onByteSlice[byteIndexInSlice]
	byteToSetBitOn, err := setBit(bit, byteToSetBitOn, bitIndexInByte)
	if err != nil {
		return onByteSlice, errors.New(fmt.Sprintf("E#ATZC9 - Can't set bit on Byte: %v", err))
	}

	// Set the byte back
	onByteSlice[byteIndexInSlice] = byteToSetBitOn

	return onByteSlice, nil
}

// / getBit gets the value of a the bit at atPosition index from the fromByte
func getBit(fromByte byte, atPosition uint16) (bool, error) {
	if atPosition > 7 || atPosition < 0 {
		return false, errors.New("E#9NHL1 - only bits 0-7 are supported by this function")
	}

	var mask byte

	switch atPosition {
	case 0:
		mask = mask0
	case 1:
		mask = mask1
	case 2:
		mask = mask2
	case 3:
		mask = mask3
	case 4:
		mask = mask4
	case 5:
		mask = mask5
	case 6:
		mask = mask6
	case 7:
		mask = mask7
	default:
		errMsg := fmt.Sprintf("E#ATB0D - INVALID POSITION: %v", atPosition)
		log.Println(errMsg)
		return false, errors.New(errMsg)
	}

	byt := fromByte

	result := byt & mask

	if result != 0 {
		// The result was not 0 so the bit in the input was set to 1
		return true, nil
	}

	return false, nil
}

// setBit sets a bit to 0 or 1 for a byte, at a given position
// The bit to be set is expressed as a boolean - true means 1, false means 0
func setBit(bit bool, onByte byte, atPosition uint16) (byte, error) {
	if atPosition > 7 || atPosition < 0 {
		return emptyByte, errors.New("E#9NHNL - only bits 0-7 are supported by this function")
	}

	var mask byte
	byt := onByte
	var result byte

	if bit {
		switch atPosition {
		case 0:
			mask = mask0
		case 1:
			mask = mask1
		case 2:
			mask = mask2
		case 3:
			mask = mask3
		case 4:
			mask = mask4
		case 5:
			mask = mask5
		case 6:
			mask = mask6
		case 7:
			mask = mask7
		default:
			errMsg := fmt.Sprintf("E#AT6RW - INVALID POSITION: %v", atPosition)
			log.Println(errMsg)
			return emptyByte, errors.New(errMsg)
		}

		result = byt | mask
	} else {
		switch atPosition {
		case 0:
			mask = clearMask0
		case 1:
			mask = clearMask1
		case 2:
			mask = clearMask2
		case 3:
			mask = clearMask3
		case 4:
			mask = clearMask4
		case 5:
			mask = clearMask5
		case 6:
			mask = clearMask6
		case 7:
			mask = clearMask7
		default:
			errMsg := fmt.Sprintf("E#ATAZN - INVALID POSITION: %v", atPosition)
			log.Println(errMsg)
			return emptyByte, errors.New(errMsg)
		}

		result = byt & mask
	}

	return result, nil
}
