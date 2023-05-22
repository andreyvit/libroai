package m

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/slices"
)

type ItemType int

const (
	ItemTypeNone             = ItemType(0)
	ItemTypeTextGeneral      = ItemType(1)
	ItemTypeTextFAQ          = ItemType(2)
	ItemTypeTextBook         = ItemType(3)
	ItemTypeTextStory        = ItemType(4)
	ItemTypeTextSummary      = ItemType(5)
	ItemTypeTextTranscript   = ItemType(6)
	ItemTypeTextInstruction  = ItemType(7)
	ItemTypeTextReserved8    = ItemType(8)
	ItemTypeTextReserved9    = ItemType(9)
	ItemTypeTextReserved10   = ItemType(10)
	ItemTypeTextReserved11   = ItemType(11)
	ItemTypeTextReserved12   = ItemType(12)
	ItemTypeTextReserved13   = ItemType(13)
	ItemTypeTextReserved14   = ItemType(14)
	ItemTypeTextReserved15   = ItemType(15)
	ItemTypeVideoGeneral     = ItemType(16)
	ItemTypeVideoQA          = ItemType(17)
	ItemTypeVideoInstruction = ItemType(18)
	ItemTypeVideoReserved19  = ItemType(19)
	ItemTypeVideoReserved20  = ItemType(20)
	ItemTypeVideoReserved21  = ItemType(21)
	ItemTypeVideoReserved22  = ItemType(22)
	ItemTypeVideoReserved23  = ItemType(23)
	ItemTypeVideoReserved24  = ItemType(24)
	ItemTypeVideoReserved25  = ItemType(25)
	ItemTypeVideoReserved26  = ItemType(26)
	ItemTypeVideoReserved27  = ItemType(27)
	ItemTypeVideoReserved28  = ItemType(28)
	ItemTypeVideoReserved29  = ItemType(29)
	ItemTypeVideoReserved30  = ItemType(30)
	ItemTypeVideoReserved31  = ItemType(31)
	ItemTypeLinkGeneral      = ItemType(32)
	ItemTypeLinkCourse       = ItemType(33)
	ItemTypeLinkTool         = ItemType(34)
	ItemTypeLinkExtraReading = ItemType(35)
	ItemTypeLinkReserved36   = ItemType(36)
	ItemTypeLinkReserved37   = ItemType(37)
	ItemTypeLinkReserved38   = ItemType(38)
	ItemTypeLinkReserved39   = ItemType(39)
	ItemTypeLinkReserved40   = ItemType(40)
	ItemTypeLinkReserved41   = ItemType(41)
	ItemTypeLinkReserved42   = ItemType(42)
	ItemTypeLinkReserved43   = ItemType(43)
	ItemTypeLinkReserved44   = ItemType(44)
	ItemTypeLinkReserved45   = ItemType(45)
	ItemTypeLinkReserved46   = ItemType(46)
	ItemTypeLinkReserved47   = ItemType(47)
)

var _itemTypeStrings = []string{
	"none",
	"text.general",
	"text.faq",
	"text.book",
	"text.story",
	"text.summary",
	"text.transcript",
	"text.instruction",
	"text.8",
	"text.9",
	"text.10",
	"text.11",
	"text.12",
	"text.13",
	"text.14",
	"text.15",
	"video.general",
	"video.qa",
	"video.instruction",
	"video.19",
	"video.20",
	"video.21",
	"video.22",
	"video.23",
	"video.24",
	"video.25",
	"video.26",
	"video.27",
	"video.28",
	"video.29",
	"video.30",
	"video.31",
	"link.general",
	"link.course",
	"link.tool",
	"link.extra",
	"link.36",
	"link.37",
	"link.38",
	"link.39",
	"link.40",
	"link.41",
	"link.42",
	"link.43",
	"link.44",
	"link.45",
	"link.46",
	"link.47",
}

func (v ItemType) String() string {
	return _itemTypeStrings[v]
}

func ParseItemType(s string) (ItemType, error) {
	if i := slices.Index(_itemTypeStrings, s); i >= 0 {
		return ItemType(i), nil
	} else {
		return ItemTypeNone, fmt.Errorf("invalid ItemType %q", s)
	}
}

func (v ItemType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *ItemType) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParseItemType(string(b))
	return err
}
func (v ItemType) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeUint(uint64(v))
}
func (v *ItemType) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeUint()
	*v = ItemType(n)
	return err
}
