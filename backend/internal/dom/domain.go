package dom

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	SurahNumber1   SurahNumber = 1
	SurahNumber2   SurahNumber = 2
	SurahNumber3   SurahNumber = 3
	SurahNumber4   SurahNumber = 4
	SurahNumber5   SurahNumber = 5
	SurahNumber6   SurahNumber = 6
	SurahNumber7   SurahNumber = 7
	SurahNumber8   SurahNumber = 8
	SurahNumber9   SurahNumber = 9
	SurahNumber10  SurahNumber = 10
	SurahNumber11  SurahNumber = 11
	SurahNumber12  SurahNumber = 12
	SurahNumber13  SurahNumber = 13
	SurahNumber14  SurahNumber = 14
	SurahNumber15  SurahNumber = 15
	SurahNumber16  SurahNumber = 16
	SurahNumber17  SurahNumber = 17
	SurahNumber18  SurahNumber = 18
	SurahNumber19  SurahNumber = 19
	SurahNumber20  SurahNumber = 20
	SurahNumber21  SurahNumber = 21
	SurahNumber22  SurahNumber = 22
	SurahNumber23  SurahNumber = 23
	SurahNumber24  SurahNumber = 24
	SurahNumber25  SurahNumber = 25
	SurahNumber26  SurahNumber = 26
	SurahNumber27  SurahNumber = 27
	SurahNumber28  SurahNumber = 28
	SurahNumber29  SurahNumber = 29
	SurahNumber30  SurahNumber = 30
	SurahNumber31  SurahNumber = 31
	SurahNumber32  SurahNumber = 32
	SurahNumber33  SurahNumber = 33
	SurahNumber34  SurahNumber = 34
	SurahNumber35  SurahNumber = 35
	SurahNumber36  SurahNumber = 36
	SurahNumber37  SurahNumber = 37
	SurahNumber38  SurahNumber = 38
	SurahNumber39  SurahNumber = 39
	SurahNumber40  SurahNumber = 40
	SurahNumber41  SurahNumber = 41
	SurahNumber42  SurahNumber = 42
	SurahNumber43  SurahNumber = 43
	SurahNumber44  SurahNumber = 44
	SurahNumber45  SurahNumber = 45
	SurahNumber46  SurahNumber = 46
	SurahNumber47  SurahNumber = 47
	SurahNumber48  SurahNumber = 48
	SurahNumber49  SurahNumber = 49
	SurahNumber50  SurahNumber = 50
	SurahNumber51  SurahNumber = 51
	SurahNumber52  SurahNumber = 52
	SurahNumber53  SurahNumber = 53
	SurahNumber54  SurahNumber = 54
	SurahNumber55  SurahNumber = 55
	SurahNumber56  SurahNumber = 56
	SurahNumber57  SurahNumber = 57
	SurahNumber58  SurahNumber = 58
	SurahNumber59  SurahNumber = 59
	SurahNumber60  SurahNumber = 60
	SurahNumber61  SurahNumber = 61
	SurahNumber62  SurahNumber = 62
	SurahNumber63  SurahNumber = 63
	SurahNumber64  SurahNumber = 64
	SurahNumber65  SurahNumber = 65
	SurahNumber66  SurahNumber = 66
	SurahNumber67  SurahNumber = 67
	SurahNumber68  SurahNumber = 68
	SurahNumber69  SurahNumber = 69
	SurahNumber70  SurahNumber = 70
	SurahNumber71  SurahNumber = 71
	SurahNumber72  SurahNumber = 72
	SurahNumber73  SurahNumber = 73
	SurahNumber74  SurahNumber = 74
	SurahNumber75  SurahNumber = 75
	SurahNumber76  SurahNumber = 76
	SurahNumber77  SurahNumber = 77
	SurahNumber78  SurahNumber = 78
	SurahNumber79  SurahNumber = 79
	SurahNumber80  SurahNumber = 80
	SurahNumber81  SurahNumber = 81
	SurahNumber82  SurahNumber = 82
	SurahNumber83  SurahNumber = 83
	SurahNumber84  SurahNumber = 84
	SurahNumber85  SurahNumber = 85
	SurahNumber86  SurahNumber = 86
	SurahNumber87  SurahNumber = 87
	SurahNumber88  SurahNumber = 88
	SurahNumber89  SurahNumber = 89
	SurahNumber90  SurahNumber = 90
	SurahNumber91  SurahNumber = 91
	SurahNumber92  SurahNumber = 92
	SurahNumber93  SurahNumber = 93
	SurahNumber94  SurahNumber = 94
	SurahNumber95  SurahNumber = 95
	SurahNumber96  SurahNumber = 96
	SurahNumber97  SurahNumber = 97
	SurahNumber98  SurahNumber = 98
	SurahNumber99  SurahNumber = 99
	SurahNumber100 SurahNumber = 100
	SurahNumber101 SurahNumber = 101
	SurahNumber102 SurahNumber = 102
	SurahNumber103 SurahNumber = 103
	SurahNumber104 SurahNumber = 104
	SurahNumber105 SurahNumber = 105
	SurahNumber106 SurahNumber = 106
	SurahNumber107 SurahNumber = 107
	SurahNumber108 SurahNumber = 108
	SurahNumber109 SurahNumber = 109
	SurahNumber110 SurahNumber = 110
	SurahNumber111 SurahNumber = 111
	SurahNumber112 SurahNumber = 112
	SurahNumber113 SurahNumber = 113
	SurahNumber114 SurahNumber = 114
	AyahNumber1    AyahNumber  = 1
	AyahNumber2    AyahNumber  = 2
	AyahNumber3    AyahNumber  = 3
	AyahNumber4    AyahNumber  = 4
	AyahNumber5    AyahNumber  = 5
	AyahNumber6    AyahNumber  = 6
	AyahNumber7    AyahNumber  = 7
	AyahNumber8    AyahNumber  = 8
	AyahNumber9    AyahNumber  = 9
	AyahNumber10   AyahNumber  = 10
	AyahNumber11   AyahNumber  = 11
	AyahNumber12   AyahNumber  = 12
	AyahNumber13   AyahNumber  = 13
	AyahNumber14   AyahNumber  = 14
	AyahNumber15   AyahNumber  = 15
	AyahNumber16   AyahNumber  = 16
	AyahNumber17   AyahNumber  = 17
	AyahNumber18   AyahNumber  = 18
	AyahNumber19   AyahNumber  = 19
	AyahNumber20   AyahNumber  = 20
	AyahNumber21   AyahNumber  = 21
	AyahNumber22   AyahNumber  = 22
	AyahNumber23   AyahNumber  = 23
	AyahNumber24   AyahNumber  = 24
	AyahNumber25   AyahNumber  = 25
	AyahNumber26   AyahNumber  = 26
	AyahNumber27   AyahNumber  = 27
	AyahNumber28   AyahNumber  = 28
	AyahNumber29   AyahNumber  = 29
	AyahNumber30   AyahNumber  = 30
	AyahNumber31   AyahNumber  = 31
	AyahNumber32   AyahNumber  = 32
	AyahNumber33   AyahNumber  = 33
	AyahNumber34   AyahNumber  = 34
	AyahNumber35   AyahNumber  = 35
	AyahNumber36   AyahNumber  = 36
	AyahNumber37   AyahNumber  = 37
	AyahNumber38   AyahNumber  = 38
	AyahNumber39   AyahNumber  = 39
	AyahNumber40   AyahNumber  = 40
	AyahNumber41   AyahNumber  = 41
	AyahNumber42   AyahNumber  = 42
	AyahNumber43   AyahNumber  = 43
	AyahNumber44   AyahNumber  = 44
	AyahNumber45   AyahNumber  = 45
	AyahNumber46   AyahNumber  = 46
	AyahNumber47   AyahNumber  = 47
	AyahNumber48   AyahNumber  = 48
	AyahNumber49   AyahNumber  = 49
	AyahNumber50   AyahNumber  = 50
	AyahNumber51   AyahNumber  = 51
	AyahNumber52   AyahNumber  = 52
	AyahNumber53   AyahNumber  = 53
	AyahNumber54   AyahNumber  = 54
	AyahNumber55   AyahNumber  = 55
	AyahNumber56   AyahNumber  = 56
	AyahNumber57   AyahNumber  = 57
	AyahNumber58   AyahNumber  = 58
	AyahNumber59   AyahNumber  = 59
	AyahNumber60   AyahNumber  = 60
	AyahNumber61   AyahNumber  = 61
	AyahNumber62   AyahNumber  = 62
	AyahNumber63   AyahNumber  = 63
	AyahNumber64   AyahNumber  = 64
	AyahNumber65   AyahNumber  = 65
	AyahNumber66   AyahNumber  = 66
	AyahNumber67   AyahNumber  = 67
	AyahNumber68   AyahNumber  = 68
	AyahNumber69   AyahNumber  = 69
	AyahNumber70   AyahNumber  = 70
	AyahNumber71   AyahNumber  = 71
	AyahNumber72   AyahNumber  = 72
	AyahNumber73   AyahNumber  = 73
	AyahNumber74   AyahNumber  = 74
	AyahNumber75   AyahNumber  = 75
	AyahNumber76   AyahNumber  = 76
	AyahNumber77   AyahNumber  = 77
	AyahNumber78   AyahNumber  = 78
	AyahNumber79   AyahNumber  = 79
	AyahNumber80   AyahNumber  = 80
	AyahNumber81   AyahNumber  = 81
	AyahNumber82   AyahNumber  = 82
	AyahNumber83   AyahNumber  = 83
	AyahNumber84   AyahNumber  = 84
	AyahNumber85   AyahNumber  = 85
	AyahNumber86   AyahNumber  = 86
	AyahNumber87   AyahNumber  = 87
	AyahNumber88   AyahNumber  = 88
	AyahNumber89   AyahNumber  = 89
	AyahNumber90   AyahNumber  = 90
	AyahNumber91   AyahNumber  = 91
	AyahNumber92   AyahNumber  = 92
	AyahNumber93   AyahNumber  = 93
	AyahNumber94   AyahNumber  = 94
	AyahNumber95   AyahNumber  = 95
	AyahNumber96   AyahNumber  = 96
	AyahNumber97   AyahNumber  = 97
	AyahNumber98   AyahNumber  = 98
	AyahNumber99   AyahNumber  = 99
	AyahNumber100  AyahNumber  = 100
	AyahNumber101  AyahNumber  = 101
	AyahNumber102  AyahNumber  = 102
	AyahNumber103  AyahNumber  = 103
	AyahNumber104  AyahNumber  = 104
	AyahNumber105  AyahNumber  = 105
	AyahNumber106  AyahNumber  = 106
	AyahNumber107  AyahNumber  = 107
	AyahNumber108  AyahNumber  = 108
	AyahNumber109  AyahNumber  = 109
	AyahNumber110  AyahNumber  = 110
	AyahNumber111  AyahNumber  = 111
	AyahNumber112  AyahNumber  = 112
	AyahNumber113  AyahNumber  = 113
	AyahNumber114  AyahNumber  = 114
	AyahNumber115  AyahNumber  = 115
	AyahNumber116  AyahNumber  = 116
	AyahNumber117  AyahNumber  = 117
	AyahNumber118  AyahNumber  = 118
	AyahNumber119  AyahNumber  = 119
	AyahNumber120  AyahNumber  = 120
	AyahNumber121  AyahNumber  = 121
	AyahNumber122  AyahNumber  = 122
	AyahNumber123  AyahNumber  = 123
	AyahNumber124  AyahNumber  = 124
	AyahNumber125  AyahNumber  = 125
	AyahNumber126  AyahNumber  = 126
	AyahNumber127  AyahNumber  = 127
	AyahNumber128  AyahNumber  = 128
	AyahNumber129  AyahNumber  = 129
	AyahNumber130  AyahNumber  = 130
	AyahNumber131  AyahNumber  = 131
	AyahNumber132  AyahNumber  = 132
	AyahNumber133  AyahNumber  = 133
	AyahNumber134  AyahNumber  = 134
	AyahNumber135  AyahNumber  = 135
	AyahNumber136  AyahNumber  = 136
	AyahNumber137  AyahNumber  = 137
	AyahNumber138  AyahNumber  = 138
	AyahNumber139  AyahNumber  = 139
	AyahNumber140  AyahNumber  = 140
	AyahNumber141  AyahNumber  = 141
	AyahNumber142  AyahNumber  = 142
	AyahNumber143  AyahNumber  = 143
	AyahNumber144  AyahNumber  = 144
	AyahNumber145  AyahNumber  = 145
	AyahNumber146  AyahNumber  = 146
	AyahNumber147  AyahNumber  = 147
	AyahNumber148  AyahNumber  = 148
	AyahNumber149  AyahNumber  = 149
	AyahNumber150  AyahNumber  = 150
	AyahNumber151  AyahNumber  = 151
	AyahNumber152  AyahNumber  = 152
	AyahNumber153  AyahNumber  = 153
	AyahNumber154  AyahNumber  = 154
	AyahNumber155  AyahNumber  = 155
	AyahNumber156  AyahNumber  = 156
	AyahNumber157  AyahNumber  = 157
	AyahNumber158  AyahNumber  = 158
	AyahNumber159  AyahNumber  = 159
	AyahNumber160  AyahNumber  = 160
	AyahNumber161  AyahNumber  = 161
	AyahNumber162  AyahNumber  = 162
	AyahNumber163  AyahNumber  = 163
	AyahNumber164  AyahNumber  = 164
	AyahNumber165  AyahNumber  = 165
	AyahNumber166  AyahNumber  = 166
	AyahNumber167  AyahNumber  = 167
	AyahNumber168  AyahNumber  = 168
	AyahNumber169  AyahNumber  = 169
	AyahNumber170  AyahNumber  = 170
	AyahNumber171  AyahNumber  = 171
	AyahNumber172  AyahNumber  = 172
	AyahNumber173  AyahNumber  = 173
	AyahNumber174  AyahNumber  = 174
	AyahNumber175  AyahNumber  = 175
	AyahNumber176  AyahNumber  = 176
	AyahNumber177  AyahNumber  = 177
	AyahNumber178  AyahNumber  = 178
	AyahNumber179  AyahNumber  = 179
	AyahNumber180  AyahNumber  = 180
	AyahNumber181  AyahNumber  = 181
	AyahNumber182  AyahNumber  = 182
	AyahNumber183  AyahNumber  = 183
	AyahNumber184  AyahNumber  = 184
	AyahNumber185  AyahNumber  = 185
	AyahNumber186  AyahNumber  = 186
	AyahNumber187  AyahNumber  = 187
	AyahNumber188  AyahNumber  = 188
	AyahNumber189  AyahNumber  = 189
	AyahNumber190  AyahNumber  = 190
	AyahNumber191  AyahNumber  = 191
	AyahNumber192  AyahNumber  = 192
	AyahNumber193  AyahNumber  = 193
	AyahNumber194  AyahNumber  = 194
	AyahNumber195  AyahNumber  = 195
	AyahNumber196  AyahNumber  = 196
	AyahNumber197  AyahNumber  = 197
	AyahNumber198  AyahNumber  = 198
	AyahNumber199  AyahNumber  = 199
	AyahNumber200  AyahNumber  = 200
	AyahNumber201  AyahNumber  = 201
	AyahNumber202  AyahNumber  = 202
	AyahNumber203  AyahNumber  = 203
	AyahNumber204  AyahNumber  = 204
	AyahNumber205  AyahNumber  = 205
	AyahNumber206  AyahNumber  = 206
	AyahNumber207  AyahNumber  = 207
	AyahNumber208  AyahNumber  = 208
	AyahNumber209  AyahNumber  = 209
	AyahNumber210  AyahNumber  = 210
	AyahNumber211  AyahNumber  = 211
	AyahNumber212  AyahNumber  = 212
	AyahNumber213  AyahNumber  = 213
	AyahNumber214  AyahNumber  = 214
	AyahNumber215  AyahNumber  = 215
	AyahNumber216  AyahNumber  = 216
	AyahNumber217  AyahNumber  = 217
	AyahNumber218  AyahNumber  = 218
	AyahNumber219  AyahNumber  = 219
	AyahNumber220  AyahNumber  = 220
	AyahNumber221  AyahNumber  = 221
	AyahNumber222  AyahNumber  = 222
	AyahNumber223  AyahNumber  = 223
	AyahNumber224  AyahNumber  = 224
	AyahNumber225  AyahNumber  = 225
	AyahNumber226  AyahNumber  = 226
	AyahNumber227  AyahNumber  = 227
	AyahNumber228  AyahNumber  = 228
	AyahNumber229  AyahNumber  = 229
	AyahNumber230  AyahNumber  = 230
	AyahNumber231  AyahNumber  = 231
	AyahNumber232  AyahNumber  = 232
	AyahNumber233  AyahNumber  = 233
	AyahNumber234  AyahNumber  = 234
	AyahNumber235  AyahNumber  = 235
	AyahNumber236  AyahNumber  = 236
	AyahNumber237  AyahNumber  = 237
	AyahNumber238  AyahNumber  = 238
	AyahNumber239  AyahNumber  = 239
	AyahNumber240  AyahNumber  = 240
	AyahNumber241  AyahNumber  = 241
	AyahNumber242  AyahNumber  = 242
	AyahNumber243  AyahNumber  = 243
	AyahNumber244  AyahNumber  = 244
	AyahNumber245  AyahNumber  = 245
	AyahNumber246  AyahNumber  = 246
	AyahNumber247  AyahNumber  = 247
	AyahNumber248  AyahNumber  = 248
	AyahNumber249  AyahNumber  = 249
	AyahNumber250  AyahNumber  = 250
	AyahNumber251  AyahNumber  = 251
	AyahNumber252  AyahNumber  = 252
	AyahNumber253  AyahNumber  = 253
	AyahNumber254  AyahNumber  = 254
	AyahNumber255  AyahNumber  = 255
	AyahNumber256  AyahNumber  = 256
	AyahNumber257  AyahNumber  = 257
	AyahNumber258  AyahNumber  = 258
	AyahNumber259  AyahNumber  = 259
	AyahNumber260  AyahNumber  = 260
	AyahNumber261  AyahNumber  = 261
	AyahNumber262  AyahNumber  = 262
	AyahNumber263  AyahNumber  = 263
	AyahNumber264  AyahNumber  = 264
	AyahNumber265  AyahNumber  = 265
	AyahNumber266  AyahNumber  = 266
	AyahNumber267  AyahNumber  = 267
	AyahNumber268  AyahNumber  = 268
	AyahNumber269  AyahNumber  = 269
	AyahNumber270  AyahNumber  = 270
	AyahNumber271  AyahNumber  = 271
	AyahNumber272  AyahNumber  = 272
	AyahNumber273  AyahNumber  = 273
	AyahNumber274  AyahNumber  = 274
	AyahNumber275  AyahNumber  = 275
	AyahNumber276  AyahNumber  = 276
	AyahNumber277  AyahNumber  = 277
	AyahNumber278  AyahNumber  = 278
	AyahNumber279  AyahNumber  = 279
	AyahNumber280  AyahNumber  = 280
	AyahNumber281  AyahNumber  = 281
	AyahNumber282  AyahNumber  = 282
	AyahNumber283  AyahNumber  = 283
	AyahNumber284  AyahNumber  = 284
	AyahNumber285  AyahNumber  = 285
	AyahNumber286  AyahNumber  = 286
)

type SurahNumber int
type AyahNumber int

type ContentType string
type Source string

const (
	ContentTypeTafsir     ContentType = "tafsir"
	SourceTafsirIbnKathir Source      = "Tafsir Ibn Kathir"
)

type MessageRole string
type LargeLanguageModel string

const (
	UserRole            MessageRole        = "user"
	ModelRole           MessageRole        = "model"
	FunctionRole        MessageRole        = "function"
	GeminiV2p5Flash     LargeLanguageModel = "gemini-2.5-flash"
	GeminiV2p5FlashLite LargeLanguageModel = "gemini-2.5-flash-lite"
)

const (
	Top5Documents  TopK = 5
	Top10Documents TopK = 10
	Top15Documents TopK = 15
	Top20Documents TopK = 20
)

type TopK int32

type Vector []float32
type rankedLists [][]int32

const MaxSubqueries int = 3
const InitialChunks200 int = 200
const RRFConstant int = 60

const (
	SurahNumberToLabelOffset int16 = 1000
	AyahNumberToLabelOffset  int16 = 2000
)

const (
	LabelContentTypeTafsir     LabelContentType = 1
	LabelSourceTafsirIbnKathir LabelSource      = 101
)

type LabelContentType int16
type LabelSource int16
type LabelSurahNumber int16
type LabelAyahNumber int16

var ContentTypeToLabel = map[ContentType]LabelContentType{
	ContentTypeTafsir: LabelContentTypeTafsir,
}

var SourceToLabel = map[Source]LabelSource{
	SourceTafsirIbnKathir: LabelSourceTafsirIbnKathir,
}

func RawToContentTypes(raw any) []ContentType {
	if raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]ContentType, 0, len(items))
	for _, v := range items {
		if s, ok := v.(string); ok {
			out = append(out, ContentType(s))
		}
	}
	return out
}

func RawToSources(raw any) []Source {
	if raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]Source, 0, len(items))
	for _, v := range items {
		if s, ok := v.(string); ok {
			out = append(out, Source(s))
		}
	}
	return out
}

func RawToSurahNumbers(raw any) []SurahNumber {
	if raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]SurahNumber, 0, len(items))
	for _, v := range items {
		switch n := v.(type) {
		case int:
			out = append(out, SurahNumber(n))
		case int32:
			out = append(out, SurahNumber(n))
		case int64:
			out = append(out, SurahNumber(n))
		case float64:
			out = append(out, SurahNumber(int(n)))
		}
	}
	return out
}

func RawToAyahNumbers(raw any) []AyahNumber {
	if raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]AyahNumber, 0, len(items))
	for _, v := range items {
		switch n := v.(type) {
		case int:
			out = append(out, AyahNumber(n))
		case int32:
			out = append(out, AyahNumber(n))
		case int64:
			out = append(out, AyahNumber(n))
		case float64:
			out = append(out, AyahNumber(int(n)))
		}
	}
	return out
}

type Document struct {
	ID          int32
	Source      Source
	Content     string
	SurahNumber SurahNumber
	AyahNumber  AyahNumber
}

type Chunk struct {
	Document
	ParentID int32
}

type User struct {
	ID                     uuid.UUID
	Email                  string
	UpdatedAt              time.Time
	TotalMessages          int32
	TotalMessagesMemorized int32
}

type Session struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	LastAccessed      time.Time
	MaxTurn           int32
	MaxTurnSummarized int32
	EndedAt           *time.Time
	Summary           *string
}

type MsgMeta struct {
	ID                int32
	SessionID         uuid.UUID
	UserID            uuid.UUID
	Turn              int32
	Model             *LargeLanguageModel
	TotalInputTokens  *int32
	TotalOutputTokens *int32
	Content           *string
	FunctionName      *string
	FunctionCall      json.RawMessage
	FunctionResponse  json.RawMessage
}

type Message interface {
	Role() MessageRole
	Meta() *MsgMeta
}

type UserMessage struct {
	MsgMeta
	MsgContent string
}

func (m *UserMessage) Role() MessageRole { return UserRole }
func (m *UserMessage) Meta() *MsgMeta    { return &m.MsgMeta }

type ModelMessage struct {
	MsgMeta
	MsgContent string
}

func (m *ModelMessage) Role() MessageRole { return ModelRole }
func (m *ModelMessage) Meta() *MsgMeta    { return &m.MsgMeta }

type FunctionMessage struct {
	MsgMeta
	FunctionName     string
	FunctionCall     json.RawMessage
	FunctionResponse json.RawMessage
}

func (m *FunctionMessage) Role() MessageRole { return FunctionRole }
func (m *FunctionMessage) Meta() *MsgMeta    { return &m.MsgMeta }

type Memory struct {
	ID         int32
	UserID     uuid.UUID
	UpdatedAt  time.Time
	SourceMsg  string
	Confidence float32
	UniqueKey  string
	Content    string
}

type LLMRole string

const (
	LLMUserRole  LLMRole = "user"
	LLMModelRole LLMRole = "model"
)

type LLMFunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

type LLMFunctionResponse struct {
	Name    string         `json:"name"`
	Content map[string]any `json:"content"`
}

type LLMPart struct {
	Text             string
	FunctionCall     *LLMFunctionCall
	FunctionResponse *LLMFunctionResponse
}

type LLMContent struct {
	Role  LLMRole
	Parts []*LLMPart
}

type LLMSchemaType string

const (
	SchemaString  LLMSchemaType = "STRING"
	SchemaInteger LLMSchemaType = "INTEGER"
	SchemaNumber  LLMSchemaType = "NUMBER"
	SchemaBoolean LLMSchemaType = "BOOLEAN"
	SchemaArray   LLMSchemaType = "ARRAY"
	SchemaObject  LLMSchemaType = "OBJECT"
)

type LLMSchema struct {
	Title       string        `json:"title,omitempty"`
	Description string        `json:"description,omitempty"`
	Type        LLMSchemaType `json:"type,omitempty"`

	Format string   `json:"format,omitempty"`
	Enum   []string `json:"enum,omitempty"`

	Required   []string              `json:"required,omitempty"`
	Properties map[string]*LLMSchema `json:"properties,omitempty"`

	Items    *LLMSchema `json:"items,omitempty"`
	MinItems *int64     `json:"minItems,omitempty"`
	MaxItems *int64     `json:"maxItems,omitempty"`

	Minimum *float64 `json:"minimum,omitempty"`
	Maximum *float64 `json:"maximum,omitempty"`

	Example any `json:"example,omitempty"`
}

type LLMFunctionDecl struct {
	Name        string
	Description string
	Parameters  *LLMSchema
}

type LLMGenConfig struct {
	SystemInstructions *LLMContent
	Temperature        float32
	CandidateCount     int32
	Tools              []*LLMFunctionDecl
	ResponseMimeType   LLMResponseSchema
	ResponseSchema     *LLMSchema
}

type LLMCountConfig struct {
	System *LLMContent
	Tools  []*LLMFunctionDecl
}

type LLMResponseSchema string

const (
	ResponseJson LLMResponseSchema = "application/json"
	ResponseText LLMResponseSchema = "text/plain"
)

func Ptr[T any](v T) *T { return &v }

func StringEnum(options ...string) *LLMSchema {
	return &LLMSchema{
		Type: SchemaString,
		Enum: options,
	}
}

func ArrayOf(item *LLMSchema, min, max *int64) *LLMSchema {
	return &LLMSchema{
		Type:     SchemaArray,
		Items:    item,
		MinItems: min,
		MaxItems: max,
	}
}

func ObjectWith(props map[string]*LLMSchema, required ...string) *LLMSchema {
	return &LLMSchema{
		Type:       SchemaObject,
		Properties: props,
		Required:   required,
	}
}

func IntegerRange(min, max *float64) *LLMSchema {
	return &LLMSchema{
		Type:    SchemaInteger,
		Minimum: min,
		Maximum: max,
	}
}

func WithDocs(title *string, description *string, s *LLMSchema) *LLMSchema {
	if title != nil {
		s.Title = *title
	}

	if description != nil {
		s.Description = *description
	}

	return s
}

type AgentName string

const (
	Caller     AgentName = "Caller"
	Generator  AgentName = "Generator"
	Summarizer AgentName = "Summarizer"
	Memorizer  AgentName = "Memorizer"
)

type AgentProfile struct {
	Model  string
	Config *LLMGenConfig
}

type LLMFunctionName string

const (
	FunctionSearch LLMFunctionName = "Search()"
)

type LLMFunctions map[LLMFunctionName]LLMFunction

type TokenUsage struct {
	InputTokens  int32
	OutputTokens int32
}

type FinishReason string

const (
	FinishReasonUnspecified           FinishReason = "FINISH_REASON_UNSPECIFIED"
	FinishReasonStop                  FinishReason = "STOP"
	FinishReasonMaxTokens             FinishReason = "MAX_TOKENS"
	FinishReasonSafety                FinishReason = "SAFETY"
	FinishReasonRecitation            FinishReason = "RECITATION"
	FinishReasonLanguage              FinishReason = "LANGUAGE"
	FinishReasonOther                 FinishReason = "OTHER"
	FinishReasonBlocklist             FinishReason = "BLOCKLIST"
	FinishReasonProhibitedContent     FinishReason = "PROHIBITED_CONTENT"
	FinishReasonSPII                  FinishReason = "SPII"
	FinishReasonMalformedFunctionCall FinishReason = "MALFORMED_FUNCTION_CALL"
	FinishReasonImageSafety           FinishReason = "IMAGE_SAFETY"
	FinishReasonUnexpectedToolCall    FinishReason = "UNEXPECTED_TOOL_CALL"
)

type LLMGenResult struct {
	Output        *ModelOutput
	Usage         *TokenUsage
	FinishReason  FinishReason
	FinishMessage string
}

const (
	TokenLimit int32 = 200_000
)

var AgentToModel = map[AgentName]LargeLanguageModel{
	Caller:    GeminiV2p5Flash,
	Generator: GeminiV2p5FlashLite,
}

func ToJsonRawMessage(m map[string]any) (json.RawMessage, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

func FromJsonRawMessage(m json.RawMessage) (map[string]any, error) {
	final := make(map[string]any)
	if err := json.Unmarshal(m, &final); err != nil {
		return nil, err
	}

	return final, nil
}

func MessagesToLLMContent(msgs []Message) ([]*LLMContent, error) {
	var win []*LLMContent
	for _, m := range msgs {
		role := m.Role()
		meta := m.Meta()
		switch role {
		case UserRole:
			win = append(win, &LLMContent{
				Role:  LLMUserRole,
				Parts: []*LLMPart{{Text: *meta.Content}},
			})
		case FunctionRole:
			call, err := FromJsonRawMessage(meta.FunctionCall)
			if err != nil {
				return nil, err
			}
			fnres, err := FromJsonRawMessage(meta.FunctionResponse)
			if err != nil {
				return nil, err
			}
			win = append(win, &LLMContent{
				Role: LLMModelRole,
				Parts: []*LLMPart{{FunctionCall: &LLMFunctionCall{
					Name: *meta.FunctionName,
					Args: call,
				}}},
			})
			win = append(win, &LLMContent{
				Role: LLMUserRole,
				Parts: []*LLMPart{{FunctionResponse: &LLMFunctionResponse{
					Name:    *meta.FunctionName,
					Content: fnres,
				}}},
			})
		case ModelRole:
			win = append(win, &LLMContent{
				Role:  LLMModelRole,
				Parts: []*LLMPart{{Text: *meta.Content}},
			})
		}
	}

	return win, nil
}

func MemoriesToLLMContent(mems []Memory) ([]*LLMContent, error) {
	final := make([]*LLMContent, 0, len(mems))
	for _, m := range mems {
		text := fmt.Sprintf(
			"Memory\n- Unique Key: %s\n- Confidence: %.2f\n- Content: %s\n- Source Msg: %s\n",
			m.UniqueKey,
			m.Confidence,
			m.Content,
			m.SourceMsg,
		)

		final = append(final, &LLMContent{
			Role: LLMUserRole,
			Parts: []*LLMPart{
				{Text: text},
			},
		})
	}
	return final, nil
}
