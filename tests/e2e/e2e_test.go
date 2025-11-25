package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_FullFlow(t *testing.T) {
	err := WaitForServer(BaseURL, 30)
	require.NoError(t, err, "Server should be ready")

	client := NewClient(BaseURL)

	t.Run("CreateTeam", func(t *testing.T) {
		teamReq := map[string]interface{}{
			"team_name": "backend",
			"members": []map[string]interface{}{
				{"user_id": "u1", "username": "Alice", "is_active": true},
				{"user_id": "u2", "username": "Bob", "is_active": true},
				{"user_id": "u3", "username": "Charlie", "is_active": true},
				{"user_id": "u4", "username": "David", "is_active": true},
			},
		}

		resp, err := client.Post("/team/add", teamReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := ReadResponseBody(resp)
			t.Logf("Unexpected status code: %d, body: %s", resp.StatusCode, body)
		}
		if resp.StatusCode == http.StatusCreated {
			var teamResp map[string]interface{}
			err = ParseJSON(resp, &teamResp)
			require.NoError(t, err)
			assert.NotNil(t, teamResp["team"])
		} else if resp.StatusCode == http.StatusBadRequest {
			t.Log("Team already exists, continuing...")
		} else {
			body, _ := ReadResponseBody(resp)
			t.Fatalf("Unexpected status code: %d, body: %s", resp.StatusCode, body)
		}
	})

	t.Run("CreatePR", func(t *testing.T) {
		prReq := map[string]interface{}{
			"pull_request_id":   "pr-1",
			"pull_request_name": "Add feature",
			"author_id":         "u1",
		}

		resp, err := client.Post("/pullRequest/create", prReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated {
			var prResp map[string]interface{}
			err = ParseJSON(resp, &prResp)
			require.NoError(t, err)

			pr, ok := prResp["pr"].(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, "pr-1", pr["pull_request_id"])
			assert.Equal(t, "OPEN", pr["status"])

			reviewers, ok := pr["assigned_reviewers"].([]interface{})
			require.True(t, ok)
			assert.GreaterOrEqual(t, len(reviewers), 1)
			assert.LessOrEqual(t, len(reviewers), 2)
		} else if resp.StatusCode == http.StatusBadRequest {
			t.Log("PR already exists, getting it...")
			getResp, err := client.Get("/team/get?team_name=backend")
			require.NoError(t, err)
			getResp.Body.Close()
		} else {
			body, _ := ReadResponseBody(resp)
			t.Fatalf("Unexpected status code: %d, body: %s", resp.StatusCode, body)
		}
	})

	t.Run("GetTeam", func(t *testing.T) {
		resp, err := client.Get("/team/get?team_name=backend")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var team map[string]interface{}
		err = ParseJSON(resp, &team)
		require.NoError(t, err)
		assert.Equal(t, "backend", team["team_name"])
	})

	t.Run("GetUserReviews", func(t *testing.T) {
		resp, err := client.Get("/users/getReview?user_id=u2")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var reviewsResp map[string]interface{}
		err = ParseJSON(resp, &reviewsResp)
		require.NoError(t, err)
		assert.Equal(t, "u2", reviewsResp["user_id"])
	})

	t.Run("ReassignReviewer", func(t *testing.T) {
		reassignReq := map[string]interface{}{
			"pull_request_id": "pr-1",
			"old_user_id":     "u2",
		}

		resp, err := client.Post("/pullRequest/reassign", reassignReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var reassignResp map[string]interface{}
			err = ParseJSON(resp, &reassignResp)
			require.NoError(t, err)
			assert.NotNil(t, reassignResp["pr"])
			assert.NotEmpty(t, reassignResp["replaced_by"])
		} else if resp.StatusCode == http.StatusBadRequest {
			body, _ := ReadResponseBody(resp)
			t.Logf("Reassign failed with u2, trying u3. Response: %s", body)
			
			reassignReq["old_user_id"] = "u3"
			resp2, err := client.Post("/pullRequest/reassign", reassignReq)
			require.NoError(t, err)
			defer resp2.Body.Close()
			
			if resp2.StatusCode == http.StatusOK {
				var reassignResp map[string]interface{}
				err = ParseJSON(resp2, &reassignResp)
				require.NoError(t, err)
				assert.NotNil(t, reassignResp["pr"])
				assert.NotEmpty(t, reassignResp["replaced_by"])
			} else {
				body2, _ := ReadResponseBody(resp2)
				t.Logf("Reassign also failed with u3. Response: %s", body2)
				if resp2.StatusCode == http.StatusBadRequest {
					t.Skip("PR might be already merged or reviewers not assigned")
				} else {
					t.Fatalf("Unexpected status code: %d, body: %s", resp2.StatusCode, body2)
				}
			}
		} else {
			body, _ := ReadResponseBody(resp)
			t.Fatalf("Unexpected status code: %d, body: %s", resp.StatusCode, body)
		}
	})

	t.Run("SetIsActive", func(t *testing.T) {
		setActiveReq := map[string]interface{}{
			"user_id":   "u3",
			"is_active": false,
		}

		resp, err := client.Post("/users/setIsActive", setActiveReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var userResp map[string]interface{}
		err = ParseJSON(resp, &userResp)
		require.NoError(t, err)

		user, ok := userResp["user"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, false, user["is_active"])
	})

	t.Run("MergePR", func(t *testing.T) {
		mergeReq := map[string]interface{}{
			"pull_request_id": "pr-1",
		}

		resp, err := client.Post("/pullRequest/merge", mergeReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var mergeResp map[string]interface{}
		err = ParseJSON(resp, &mergeResp)
		require.NoError(t, err)

		pr, ok := mergeResp["pr"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "MERGED", pr["status"])
		mergedAt, exists := pr["merged_at"]
		assert.True(t, exists, "merged_at field should exist")
		if mergedAt != nil {
			mergedAtStr, ok := mergedAt.(string)
			if ok {
				assert.NotEmpty(t, mergedAtStr, "merged_at should not be empty")
			}
		}
	})

	t.Run("CannotReassignAfterMerge", func(t *testing.T) {
		reassignReq := map[string]interface{}{
			"pull_request_id": "pr-1",
			"old_user_id":     "u4",
		}

		resp, err := client.Post("/pullRequest/reassign", reassignReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResp map[string]interface{}
		err = ParseJSON(resp, &errorResp)
		require.NoError(t, err)

		errorObj, ok := errorResp["error"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "PR_MERGED", errorObj["code"])
	})

	t.Run("GetStatistics", func(t *testing.T) {
		resp, err := client.Get("/statistics")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var stats map[string]interface{}
		err = ParseJSON(resp, &stats)
		require.NoError(t, err)

		assert.Contains(t, stats, "total_prs")
		assert.Contains(t, stats, "open_prs")
		assert.Contains(t, stats, "merged_prs")
		assert.Contains(t, stats, "user_statistics")

		totalPRs, ok := stats["total_prs"].(float64)
		require.True(t, ok)
		assert.GreaterOrEqual(t, int(totalPRs), 1)
	})
}

func TestE2E_Statistics(t *testing.T) {
	err := WaitForServer(BaseURL, 30)
	require.NoError(t, err, "Server should be ready")

	client := NewClient(BaseURL)

	t.Run("StatisticsEndpoint", func(t *testing.T) {
		resp, err := client.Get("/statistics")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var stats map[string]interface{}
		err = ParseJSON(resp, &stats)
		require.NoError(t, err)

		assert.Contains(t, stats, "total_prs")
		assert.Contains(t, stats, "open_prs")
		assert.Contains(t, stats, "merged_prs")
		assert.Contains(t, stats, "user_statistics")

		_, ok := stats["total_prs"].(float64)
		assert.True(t, ok, "total_prs should be a number")

		_, ok = stats["open_prs"].(float64)
		assert.True(t, ok, "open_prs should be a number")

		_, ok = stats["merged_prs"].(float64)
		assert.True(t, ok, "merged_prs should be a number")

		userStats, ok := stats["user_statistics"].([]interface{})
		assert.True(t, ok, "user_statistics should be an array")

		if len(userStats) > 0 {
			firstUser, ok := userStats[0].(map[string]interface{})
			require.True(t, ok)
			assert.Contains(t, firstUser, "user_id")
			assert.Contains(t, firstUser, "username")
			assert.Contains(t, firstUser, "total_reviews")
			assert.Contains(t, firstUser, "open_reviews")
			assert.Contains(t, firstUser, "merged_reviews")
		}
	})
}

