package workflows

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const incidentReaction = "one"
const notAnIncidentReaction = "two"

//todo tests

func Test_hasIsIncidentReactionOnMessage_shouldPassHappyPath(t *testing.T) {
	reactionKeysMap := make(map[string]bool)
	reactionKeysMap[incidentReaction] = true
	reactionCountsMap := make(map[string]int)
	reactionCountsMap[incidentReaction] = 2
	got := hasIsIncidentReactionOnMessage(reactionKeysMap, reactionCountsMap)
	assert.True(t, got)
}

func Test_hasIsIncidentReactionOnMessage_shouldFailIfOnlyHasStarterReaction(t *testing.T) {
	reactionKeysMap := make(map[string]bool)
	reactionKeysMap[incidentReaction] = true
	reactionCountsMap := make(map[string]int)
	reactionCountsMap[incidentReaction] = 1
	got := hasIsIncidentReactionOnMessage(reactionKeysMap, reactionCountsMap)
	assert.False(t, got)
}

func Test_hasIsIncidentReactionOnMessage_shouldFailIfMissingIncidentReaction(t *testing.T) {
	reactionKeysMap := make(map[string]bool)
	reactionCountsMap := make(map[string]int)
	reactionCountsMap[incidentReaction] = 2
	got := hasIsIncidentReactionOnMessage(reactionKeysMap, reactionCountsMap)
	assert.False(t, got)
}

func Test_hasNotAnIncidentReactionOnMessage_shouldPassHappyPath(t *testing.T) {
	reactionKeysMap := make(map[string]bool)
	reactionKeysMap[notAnIncidentReaction] = true
	reactionCountsMap := make(map[string]int)
	reactionCountsMap[notAnIncidentReaction] = 2
	got := hasNotAnIncidentReactionOnMessage(reactionKeysMap, reactionCountsMap)
	assert.True(t, got)
}

func Test_hasNotAnIncidentReactionOnMessage_shouldFailIfOnlyHasStarterReaction(t *testing.T) {
	reactionKeysMap := make(map[string]bool)
	reactionKeysMap[notAnIncidentReaction] = true
	reactionCountsMap := make(map[string]int)
	reactionCountsMap[notAnIncidentReaction] = 1
	got := hasNotAnIncidentReactionOnMessage(reactionKeysMap, reactionCountsMap)
	assert.False(t, got)
}

func Test_hasNotAnIncidentReactionOnMessage_shouldFailIfMissingIncidentReaction(t *testing.T) {
	reactionKeysMap := make(map[string]bool)
	reactionCountsMap := make(map[string]int)
	reactionCountsMap[notAnIncidentReaction] = 2
	got := hasNotAnIncidentReactionOnMessage(reactionKeysMap, reactionCountsMap)
	assert.False(t, got)
}

func Test_lookupSlackData_shouldBuildSlackActivityData(t *testing.T) {

	got := generateIncidentAlertMessage("foo")
	assert.NotNil(t, got)
	assert.Equal(t, "fooIt looks like there might be an error. \n:one: To confirm the incident and start debugging \n:two: To dismiss", got.FirstResponseWarning)
	assert.Equal(t, "Here's the stack trace.", got.Attachment.Pretext)
	assert.Equal(t, "Traceback (most recent call last):\n  File \"tb.py\", line 15, in <module>\n    a()\n  File \"tb.py\", line 3, in a\n    j = b(i)\n  File \"tb.py\", line 9, in b\n    c()\n  File \"tb.py\", line 13, in c\n    error()\nNameError: name 'error' is not defined\n", got.Attachment.Text)

}
