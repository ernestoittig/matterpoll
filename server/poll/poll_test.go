package poll_test

import (
	"testing"

	"github.com/bouk/monkey"
	"github.com/mattermost/mattermost-server/model"
	"github.com/matterpoll/matterpoll/server/poll"
	"github.com/matterpoll/matterpoll/server/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPoll(t *testing.T) {
	t.Run("all fine", func(t *testing.T) {
		assert := assert.New(t)
		patch := monkey.Patch(model.GetMillis, func() int64 { return 1234567890 })
		defer patch.Unpatch()

		creator := model.NewRandomString(10)
		question := model.NewRandomString(10)
		answerOptions := []string{model.NewRandomString(10), model.NewRandomString(10), model.NewRandomString(10)}
		p, err := poll.NewPoll("v1", creator, question, answerOptions, []string{"anonymous", "progress"})

		require.Nil(t, err)
		require.NotNil(t, p)
		assert.Equal(int64(1234567890), p.CreatedAt)
		assert.Equal(creator, p.Creator)
		assert.Equal("v1", p.DataSchemaVersion)
		assert.Equal(question, p.Question)
		assert.Equal(&poll.AnswerOption{Answer: answerOptions[0], Voter: nil}, p.AnswerOptions[0])
		assert.Equal(&poll.AnswerOption{Answer: answerOptions[1], Voter: nil}, p.AnswerOptions[1])
		assert.Equal(&poll.AnswerOption{Answer: answerOptions[2], Voter: nil}, p.AnswerOptions[2])
		assert.Equal(poll.PollSettings{Anonymous: true, Progress: true}, p.Settings)
	})
	t.Run("error", func(t *testing.T) {
		assert := assert.New(t)

		creator := model.NewRandomString(10)
		question := model.NewRandomString(10)
		answerOptions := []string{model.NewRandomString(10), model.NewRandomString(10), model.NewRandomString(10)}
		p, err := poll.NewPoll("v1", creator, question, answerOptions, []string{"unkownOption"})

		assert.Nil(p)
		assert.NotNil(err)
	})
}

func TestEncodeDecode(t *testing.T) {
	p1 := testutils.GetPollWithVotes()
	p2 := poll.DecodePollFromByte(p1.EncodeToByte())
	assert.Equal(t, p1, p2)
}

func TestDecode(t *testing.T) {
	p := poll.DecodePollFromByte([]byte{})
	assert.Nil(t, p)
}

func TestUpdateVote(t *testing.T) {
	for name, test := range map[string]struct {
		Poll         poll.Poll
		UserID       string
		Index        int
		ExpectedPoll poll.Poll
		Error        bool
	}{
		"Negative Index": {
			Poll: poll.Poll{
				Question: "Question",
				AnswerOptions: []*poll.AnswerOption{
					{Answer: "Answer 1",
						Voter: []string{"a"}},
					{Answer: "Answer 2"},
				},
			},
			UserID: "a",
			Index:  -1,
			ExpectedPoll: poll.Poll{
				Question: "Question",
				AnswerOptions: []*poll.AnswerOption{
					{Answer: "Answer 1",
						Voter: []string{"a"}},
					{Answer: "Answer 2"},
				},
			},
			Error: true,
		},
		"To high Index": {
			Poll: poll.Poll{
				Question: "Question",
				AnswerOptions: []*poll.AnswerOption{
					{Answer: "Answer 1",
						Voter: []string{"a"}},
					{Answer: "Answer 2"},
				},
			},
			UserID: "a",
			Index:  2,
			ExpectedPoll: poll.Poll{
				Question: "Question",
				AnswerOptions: []*poll.AnswerOption{
					{Answer: "Answer 1",
						Voter: []string{"a"}},
					{Answer: "Answer 2"},
				},
			},
			Error: true,
		},
		"Invalid userID": {
			Poll: poll.Poll{
				Question: "Question",
				AnswerOptions: []*poll.AnswerOption{
					{Answer: "Answer 1",
						Voter: []string{"a"}},
					{Answer: "Answer 2"},
				},
			},
			UserID: "",
			Index:  1,
			ExpectedPoll: poll.Poll{
				Question: "Question",
				AnswerOptions: []*poll.AnswerOption{
					{Answer: "Answer 1",
						Voter: []string{"a"}},
					{Answer: "Answer 2"},
				},
			},
			Error: true,
		},
		"Idempotent": {
			Poll: poll.Poll{
				Question: "Question",
				AnswerOptions: []*poll.AnswerOption{
					{Answer: "Answer 1",
						Voter: []string{"a"}},
					{Answer: "Answer 2"},
				},
			},
			UserID: "a",
			Index:  0,
			ExpectedPoll: poll.Poll{
				Question: "Question",
				AnswerOptions: []*poll.AnswerOption{
					{Answer: "Answer 1",
						Voter: []string{"a"}},
					{Answer: "Answer 2"},
				},
			},
			Error: false,
		},
		"Valid Vote": {
			Poll: poll.Poll{
				Question: "Question",
				AnswerOptions: []*poll.AnswerOption{
					{Answer: "Answer 1",
						Voter: []string{"a"}},
					{Answer: "Answer 2"},
				},
			},
			UserID: "a",
			Index:  1,
			ExpectedPoll: poll.Poll{
				Question: "Question",
				AnswerOptions: []*poll.AnswerOption{
					{Answer: "Answer 1",
						Voter: []string{}},
					{Answer: "Answer 2",
						Voter: []string{"a"}},
				},
			},
			Error: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := test.Poll.UpdateVote(test.UserID, test.Index)

			if test.Error {
				assert.NotNil(err)
			} else {
				assert.Nil(err)
			}
			assert.Equal(test.ExpectedPoll, test.Poll)
		})
	}
}

func TestHasVoted(t *testing.T) {
	p1 := &poll.Poll{Question: "Question",
		AnswerOptions: []*poll.AnswerOption{
			{Answer: "Answer 1",
				Voter: []string{"a"}},
			{Answer: "Answer 2"},
		},
	}
	assert.True(t, p1.HasVoted("a"))
	assert.False(t, p1.HasVoted("b"))
}

func TestPollCopy(t *testing.T) {
	assert := assert.New(t)

	t.Run("no change", func(t *testing.T) {
		p := testutils.GetPoll()
		p2 := p.Copy()

		assert.Equal(p, p2)
	})
	t.Run("change Question", func(t *testing.T) {
		p := testutils.GetPoll()
		p2 := p.Copy()

		p.Question = "Different question"
		assert.NotEqual(p.Question, p2.Question)
		assert.NotEqual(p, p2)
	})
	t.Run("change AnswerOptions", func(t *testing.T) {
		p := testutils.GetPoll()
		p2 := p.Copy()

		p.AnswerOptions[0].Answer = "abc"
		assert.NotEqual(p.AnswerOptions[0].Answer, p2.AnswerOptions[0].Answer)
		assert.NotEqual(p, p2)
	})
	t.Run("change Settings", func(t *testing.T) {
		p := testutils.GetPoll()
		p2 := p.Copy()

		p.Settings.Progress = true
		assert.NotEqual(p.Settings.Progress, p2.Settings.Progress)
		assert.NotEqual(p, p2)
	})
}