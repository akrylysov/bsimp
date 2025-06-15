package media

// import (
// 	// "github.com/ginqi7/bsimp/internal/storage"
// 	"github.com/stretchr/testify/assert"
// 	"testing"
// )

// func file// s(paths ...string) []*storage.StorageFile {
// 	var files []*storage.StorageFile
// 	for _, p := range paths {
// 		files = append(files, storage.NewStorageFile(p, 1))
// 	}
// 	return files
// }

// func TestScoreCovers(t *testing.T) {
// 	in := files("1.mp3", "2.JpG", "3.GIF", "cover.jpg", "abc", "1_cover.png")
// 	expected := []ScoredFile{
// 		{
// 			StorageFile: in[1],
// 			Score:       0,
// 		},
// 		{
// 			StorageFile: in[2],
// 			Score:       0,
// 		},
// 		{
// 			StorageFile: in[3],
// 			Score:       2,
// 		},
// 		{
// 			StorageFile: in[5],
// 			Score:       1,
// 		},
// 	}
// 	actual := ScoreCovers(in)
// 	assert.EqualValues(t, expected, actual)
// }

// func TestIsAudioFile(t *testing.T) {
// 	in := files("1.mp3", "abc", "cover.jpg", "2.ogg", "3.MP3")
// 	expected := []bool{true, false, false, true, true}
// 	for i, f := range in {
// 		actual := IsAudioFile(f)
// 		assert.Equal(t, expected[i], actual)
// 	}
// }
