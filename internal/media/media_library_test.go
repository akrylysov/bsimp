package media

// import (
// 	"strings"
// 	"testing"

// 	"github.com/aws/aws-sdk-go/aws"
// 	"github.com/aws/aws-sdk-go/service/s3"
// 	"github.com/ginqi7/bsimp/internal/config"
// 	"github.com/ginqi7/bsimp/internal/storage"
// 	"github.com/stretchr/testify/assert"
// )

// func TestMediaLibrary(t *testing.T) {
// asrt := assert.New(t)

// cfg, closeS3 := config.NewTestS3Config()
// defer closeS3()
// cfg.BasePrefix = "music/"
// s3Storage, err := storage.NewS3Storage(cfg)
// asrt.NoError(err)

// _, err = s3Storage.s3.CreateBucket(&s3.CreateBucketInput{
// 	Bucket: aws.String("test"),
// })
// asrt.NoError(err)

// keys := []string{
// 	"music/Aphex Twin/1992 - Selected Ambient Works 85-92/01. Xtal.mp3",
// 	"music/Aphex Twin/1992 - Selected Ambient Works 85-92/Cover.jpg",
// 	"music/Aphex Twin/1999 - Windowlicker/01 Windowlicker.mp3",
// 	"music/Aphex Twin/1999 - Windowlicker/02 [Equation].mp3",
// 	"music/Aphex Twin/1999 - Windowlicker/03 Nannou.mp3",
// 	"music/Aphex Twin/1999 - Windowlicker/Folder.jpg",
// 	"music/Aphex Twin/1999 - Windowlicker/back.jpg",
// 	"music/Aphex Twin/1999 - Windowlicker/covers/front_cover.jpg",
// 	"music/The Prodigy/1992 - The Prodigy Experience/Scans/Cover-Case.png",
// 	"music/The Prodigy/1992 - The Prodigy Experience/CD1/01 - Jericho.mp3",
// 	"music/The Prodigy/1992 - The Prodigy Experience/CD2/01 - Your Love.mp3",
// 	"music/Venetian Snares/2016 - Traditional Synthesizer Music/01. Dreamt Person v3.mp3",
// 	"music/Venetian Snares/2016 - Traditional Synthesizer Music/tracklist.txt",
// }
// for _, key := range keys {
// 	_, err := storage.s3.PutObject(&s3.PutObjectInput{
// 		Body:   strings.NewReader("1"),
// 		Bucket: aws.String("test"),
// 		Key:    aws.String(key),
// 	})
// 	asrt.NoError(err)
// }

// testCases := map[string]MediaListing{
// 	"": {
// 		CurrentDirectory: storage.NewStorageDirectory(""),
// 		Directories: []*storage.StorageDirectory{
// 			storage.NewStorageDirectory("Aphex Twin"),
// 			storage.NewStorageDirectory("The Prodigy"),
// 			storage.NewStorageDirectory("Venetian Snares"),
// 		},
// 	},
// 	"Aphex Twin": {
// 		CurrentDirectory: storage.NewStorageDirectory("Aphex Twin"),
// 		Directories: []*storage.StorageDirectory{
// 			storage.NewStorageDirectory("Aphex Twin/1992 - Selected Ambient Works 85-92"),
// 			storage.NewStorageDirectory("Aphex Twin/1999 - Windowlicker"),
// 		},
// 	},
// 	"Aphex Twin/1992 - Selected Ambient Works 85-92": {
// 		CurrentDirectory: storage.NewStorageDirectory("Aphex Twin/1992 - Selected Ambient Works 85-92"),
// 		AudioTracks: []*storage.StorageFile{
// 			NewStorageFile("Aphex Twin/1992 - Selected Ambient Works 85-92/01. Xtal.mp3", 1),
// 		},
// 		Cover: NewStorageFile("Aphex Twin/1992 - Selected Ambient Works 85-92/Cover.jpg", 1),
// 	},
// 	"Aphex Twin/1999 - Windowlicker": {
// 		CurrentDirectory: storage.NewStorageDirectory("Aphex Twin/1999 - Windowlicker"),
// 		AudioTracks: []*storage.StorageFile{
// 			NewStorageFile("Aphex Twin/1999 - Windowlicker/01 Windowlicker.mp3", 1),
// 			NewStorageFile("Aphex Twin/1999 - Windowlicker/02 [Equation].mp3", 1),
// 			NewStorageFile("Aphex Twin/1999 - Windowlicker/03 Nannou.mp3", 1),
// 		},
// 		Cover: NewStorageFile("Aphex Twin/1999 - Windowlicker/Folder.jpg", 1),
// 		Directories: []*storage.StorageDirectory{
// 			storage.NewStorageDirectory("Aphex Twin/1999 - Windowlicker/covers"),
// 		},
// 		Files: []*storage.StorageFile{
// 			NewStorageFile("Aphex Twin/1999 - Windowlicker/back.jpg", 1),
// 		},
// 	},
// 	"The Prodigy": {
// 		CurrentDirectory: storage.NewStorageDirectory("The Prodigy"),
// 		Directories: []*storage.StorageDirectory{
// 			storage.NewStorageDirectory("The Prodigy/1992 - The Prodigy Experience"),
// 		},
// 	},
// 	"The Prodigy/1992 - The Prodigy Experience": {
// 		CurrentDirectory: storage.NewStorageDirectory("The Prodigy/1992 - The Prodigy Experience"),
// 		Directories: []*storage.StorageDirectory{
// 			storage.NewStorageDirectory("The Prodigy/1992 - The Prodigy Experience/CD1"),
// 			storage.NewStorageDirectory("The Prodigy/1992 - The Prodigy Experience/CD2"),
// 			storage.NewStorageDirectory("The Prodigy/1992 - The Prodigy Experience/Scans"),
// 		},
// 		Cover: NewStorageFile("The Prodigy/1992 - The Prodigy Experience/Scans/Cover-Case.png", 1),
// 	},
// 	"The Prodigy/1992 - The Prodigy Experience/CD1": {
// 		CurrentDirectory: storage.NewStorageDirectory("The Prodigy/1992 - The Prodigy Experience/CD1"),
// 		AudioTracks: []*storage.StorageFile{
// 			NewStorageFile("The Prodigy/1992 - The Prodigy Experience/CD1/01 - Jericho.mp3", 1),
// 		},
// 	},
// 	"Venetian Snares": {
// 		CurrentDirectory: storage.NewStorageDirectory("Venetian Snares"),
// 		Directories: []*storage.StorageDirectory{
// 			storage.NewStorageDirectory("Venetian Snares/2016 - Traditional Synthesizer Music"),
// 		},
// 	},
// 	"Venetian Snares/2016 - Traditional Synthesizer Music": {
// 		CurrentDirectory: storage.NewStorageDirectory("Venetian Snares/2016 - Traditional Synthesizer Music"),
// 		AudioTracks: []*storage.StorageFile{
// 			NewStorageFile("Venetian Snares/2016 - Traditional Synthesizer Music/01. Dreamt Person v3.mp3", 1),
// 		},
// 		Files: []*storage.StorageFile{
// 			NewStorageFile("Venetian Snares/2016 - Traditional Synthesizer Music/tracklist.txt", 1),
// 		},
// 	},
// }

// ml := NewMediaLibrary(storage)
// for path, expectedListing := range testCases {
// 	l, err := ml.List(path)
// 	asrt.NoError(err)
// 	asrt.EqualValues(&expectedListing, l, path)
// }

// // Path doesn't exist.
// _, err = ml.List("music")
// asrt.Error(err)
// _, err = ml.List("The Prodigy/1992 - The Prodigy Experience/CD3")
// asrt.Error(err)
// }
