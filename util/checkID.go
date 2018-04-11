package util

import (
	"encoding/binary"
	"errors"
	"strconv"
)

//CheckArmorID 驗証 ArmorID是否正確
func CheckArmorID(armorID string) (int, error) {

	var err error
	if len(armorID) != 12 {
		err = errors.New("id size not enough")
		return 0, err
	}

	lenID := len(armorID)
	checksum, err := strconv.Atoi(armorID[lenID-2 : lenID])
	if err != nil {
		return 0, err
	}

	currentID, err := strconv.Atoi(armorID[0 : lenID-2])
	if err != nil {
		return 0, err
	}

	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(currentID))
	calsum := 0
	for i := 0; i < 4; i++ {
		calsum += int(buf[i] & 0x0f)
	}

	if calsum != checksum {
		err = errors.New("check sum error")
		return 0, err
	}

	return currentID, err
}
