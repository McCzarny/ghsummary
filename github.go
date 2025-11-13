package ghsummary

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Activity struct {
	Type       string
	Repository string
	Content    string
}

// makeGitHubRequest creates an HTTP GET request with GitHub token authentication if available
func makeGitHubRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add GitHub token if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		log.Printf("Using GitHub token for authentication")
	}

	client := &http.Client{}
	return client.Do(req)
}

func GetRepositoryName(event map[string]interface{}) (string, error) {
	repo, ok := event["repo"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("error parsing repo data")
	}
	name, ok := repo["name"].(string)
	if !ok {
		return "", fmt.Errorf("error parsing repo name")
	}
	return name, nil
}

func GetIssueCommentEventContent(payload map[string]interface{}) (string, error) {
	action, ok := payload["action"].(string)
	if !ok {
		return "", fmt.Errorf("error parsing action data")
	}
	switch action {
	case "created":
		issue, ok := payload["issue"].(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("error parsing issue data")
		}
		title, ok := issue["title"].(string)
		if !ok {
			return "", fmt.Errorf("error parsing issue title")
		}
		body, ok := issue["body"].(string)
		if !ok {
			log.Printf("No body, using empty string")
			body = ""
		}
		return fmt.Sprintf("Issue created: %s\n%s", title, body), nil
	}
	return "", fmt.Errorf("unsupported action: %s", action)
}

func GetPushEventCommits(repo string, before string, after string) ([]interface{}, error) {
	// Use GitHub's compare API to fetch commits between two SHAs
	url := fmt.Sprintf("https://api.github.com/repos/%s/compare/%s...%s", repo, before, after)
	log.Printf("Fetching commits from compare API: %s", url)

	resp, err := makeGitHubRequest(url)
	if err != nil {
		log.Printf("Error making HTTP request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Non-OK HTTP status: %s", resp.Status)
		return nil, fmt.Errorf("failed to fetch commits: %s", resp.Status)
	}

	var compareData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&compareData); err != nil {
		log.Printf("Error decoding JSON response: %v", err)
		return nil, err
	}

	commits, ok := compareData["commits"].([]interface{})
	if !ok {
		log.Printf("Error parsing commits data from compare API")
		return nil, fmt.Errorf("error parsing commits data")
	}

	return commits, nil
}

func GetCommitSummary(commit map[string]interface{}) (string, bool) {
	message, ok := commit["message"].(string)
	if !ok {
		return "", false
	}
	url, ok := commit["url"].(string)
	if !ok {
		return "", false
	}
	resp, err := makeGitHubRequest(url)
	if err != nil {
		log.Printf("Error making HTTP request: %v", err)
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Non-OK HTTP status: %s", resp.Status)
		return "", false
	}
	var commitData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&commitData); err != nil {
		log.Printf("Error decoding JSON response: %v", err)
		return "", false
	}

	commitContentToSummarize := fmt.Sprintf("Commit message: %s\n", message)
	if files, ok := commitData["files"].([]interface{}); ok {
		for _, file := range files {
			fileData, ok := file.(map[string]interface{})
			if !ok {
				log.Printf("Error parsing file data")
				continue
			}
			filename, ok := fileData["filename"].(string)
			if !ok {
				log.Printf("Error parsing filename")
				continue
			}
			patch, ok := fileData["patch"].(string)
			if !ok {
				log.Printf("Error parsing patch data for url: %s and file: %s", url, filename)
				continue
			}
			commitContentToSummarize += fmt.Sprintf("File: %s\nPatch:\n%s\n", filename, patch)
		}
	} else {
		log.Printf("Error parsing files data")
		return "", false
	}

	commit_summary, err := GenerateCommitSummary(commitContentToSummarize, 0)
	if err != nil {
		log.Printf("Error generating commit summary: %v", err)
		return "", false
	}
	if commit_summary == "" {
		log.Printf("No commit summary generated")
		return "", false
	}

	return commit_summary, true
}

func GetEvents(username string, perPageEvents int, page int) ([]map[string]interface{}, error) {
	log.Printf("Fetching %d events for %s user. Page %d", perPageEvents, username, page)
	url := fmt.Sprintf("https://api.github.com/users/%s/events?per_page=%d&page=%d", username, perPageEvents, page)
	log.Printf("Making HTTP GET request to URL: %s", url)

	// Set up HTTP client with timeout
	resp, err := makeGitHubRequest(url)
	if err != nil {
		log.Printf("Error making HTTP request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Non-OK HTTP status: %s", resp.Status)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
		} else {
			log.Printf("body: %s", body)
		}
		return nil, fmt.Errorf("failed to fetch activity: %s", resp.Status)
	}

	var events []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		log.Printf("Error decoding JSON response: %v", err)
		return nil, err
	}

	return events, nil
}

func ProcessActivities(
	events []map[string]interface{},
	maxEvents int,
	mode string,
	maxCommitSummary int,
	activities *[]Activity,
	repositories *map[string]struct{},
	commitSummariesCount *int) {
	for _, event := range events {
		if len(*activities) >= maxEvents {
			log.Printf("Reached maximum number of activities to process: %d", maxEvents)
			break
		}

		id, ok := event["id"].(string)
		if !ok {
			log.Printf("Error parsing event ID")
			continue
		}
		if eventType, ok := event["type"].(string); ok {
			log.Printf("Processing event type: %s", eventType)
			switch eventType {
			case "IssueCommentEvent":
				repo, err := GetRepositoryName(event)
				if err != nil {
					log.Printf("[%s] Error getting repository name", id)
					continue
				}
				log.Printf("[%s] Processing IssueCommentEvent for repo: %s", id, repo)
				payload, ok := event["payload"].(map[string]interface{})
				if !ok {
					log.Printf("[%s] Error parsing payload data", id)
					continue
				}
				content, err := GetIssueCommentEventContent(payload)
				if err != nil {
					log.Printf("[%s] Error getting issue comment content: %v", id, err)
					continue
				}

				if _, exists := (*repositories)[repo]; !exists {
					(*repositories)[repo] = struct{}{}
					log.Printf("[%s] Adding repository: %s", id, repo)
				}
				*activities = append(*activities, Activity{
					Type:       eventType,
					Repository: repo,
					Content:    content,
				})
			case "PushEvent":
				repo, err := GetRepositoryName(event)
				if err != nil {
					log.Printf("[%s] Error getting repository name", id)
					continue
				}
				log.Printf("[%s] Processing PushEvent for repo: %s", id, repo)
				payload, ok := event["payload"].(map[string]interface{})
				if !ok {
					log.Printf("[%s] Error parsing payload data", id)
					continue
				}

				// Get before and after SHAs from the payload
				before, ok := payload["before"].(string)
				if !ok {
					log.Printf("[%s] Error parsing 'before' SHA", id)
					continue
				}
				after, ok := payload["head"].(string)
				if !ok {
					log.Printf("[%s] Error parsing 'head' SHA", id)
					continue
				}

				// Fetch commits using the compare API
				commits, err := GetPushEventCommits(repo, before, after)
				if err != nil {
					log.Printf("[%s] Error fetching commits from compare API: %v", id, err)
					continue
				}

				messages := ""
				for _, commit := range commits {
					commitData, ok := commit.(map[string]interface{})
					if !ok {
						log.Printf("[%s] Error parsing commit data", id)
						continue
					}

					// Extract message from commit.commit.message structure
					commitInfo, ok := commitData["commit"].(map[string]interface{})
					if !ok {
						log.Printf("[%s] Error parsing commit info", id)
						continue
					}
					message, ok := commitInfo["message"].(string)
					if !ok {
						log.Printf("[%s] Error parsing commit message", id)
						continue
					}

					if strings.EqualFold(mode, "strict") && *commitSummariesCount < maxCommitSummary {
						// For strict mode, we need to create a compatible structure for GetCommitSummary
						commitForSummary := map[string]interface{}{
							"message": message,
							"url":     commitData["url"],
						}
						commit_summary, ok := GetCommitSummary(commitForSummary)
						if !ok {
							log.Printf("[%s] Error generating commit summary", id)
							messages += message + "\n"
							continue
						}

						summary := fmt.Sprintf("Commit summary: %s", commit_summary)
						messages += summary + "\n"
						(*commitSummariesCount)++
					} else {
						messages += message + "\n"
					}
				}

				if messages == "" {
					log.Printf("[%s] No commit messages found", id)
					continue
				}

				if _, exists := (*repositories)[repo]; !exists {
					(*repositories)[repo] = struct{}{}
					log.Printf("[%s] Adding repository: %s", id, repo)
				}
				*activities = append(*activities, Activity{
					Type:       eventType,
					Repository: repo,
					Content:    messages,
				})

			default:
				log.Printf("[%s] Unsupported event type: %s", id, eventType)
			}
		}
	}
}

func GetUserActivity(username string, maxEvents int, mode string) (string, error) {
	maxEvents = min(maxEvents, 100) // Limit to 100 events. Pagination is implemented below.
	maxCommitSummary := 10          // Limit the number of commit summaries as they need to be additionally processed.
	log.Printf("Fetching activity for user: %s with max events: %d in mode: %s", username, maxEvents, mode)
	minActivityCount := 10
	activities := []Activity{}
	repositories := make(map[string]struct{})
	commitSummariesCount := 0
	currentPage := 1

	const maxPagesAllowed = 4 // As of today 11.11.2025 GitHub limits the number of pages to 3 for this endpoint.
	for len(activities) < minActivityCount && currentPage < maxPagesAllowed {
		log.Printf("Fetching page %d of events for user: %s", currentPage, username)
		log.Printf("Current number of activities: %d. Minimal number for activities: %d", len(activities), minActivityCount)
		events, err := GetEvents(username, maxEvents, currentPage)

		if err != nil {
			log.Printf("Error making HTTP request: %v", err)
			return "", err
		}

		log.Printf("Successfully fetched %d events", len(events))

		ProcessActivities(events, maxEvents, mode, maxCommitSummary, &activities, &repositories, &commitSummariesCount)
		currentPage++
	}

	recentActivities := fmt.Sprintf("Recent activities for user %s:\n", username)
	recentActivities += "Information about the repositories:\n"
	for repo := range repositories {
		readmeURL := fmt.Sprintf("https://api.github.com/repos/%s/contents/README.md", repo)
		resp, err := makeGitHubRequest(readmeURL)
		if err != nil {
			log.Printf("Error fetching README.md for repo %s: %v", repo, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			response, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading response body for repo %s: %v", repo, err)
				continue
			}

			var readmeContent map[string]interface{}
			if err := json.Unmarshal(response, &readmeContent); err != nil {
				log.Printf("Error decoding JSON response for repo %s: %v", repo, err)
				continue
			}
			content, ok := readmeContent["content"].(string)
			if !ok {
				log.Printf("Error parsing README.md content for repo %s", repo)
				continue
			}

			// Decode the base64 content
			decodedContent, err := base64.StdEncoding.DecodeString(content)
			if err != nil {
				log.Printf("Error decoding base64 content for repo %s: %v", repo, err)
				continue
			}

			recentActivities += fmt.Sprintf("%s repository description:\n%s\n\n", repo, decodedContent)
		} else {
			log.Printf("No README.md found for repo %s", repo)
		}
	}

	for _, activity := range activities {
		recentActivities += fmt.Sprintf("Type: %s\nRepository: %s\nContent: %s\n\n", activity.Type, activity.Repository, activity.Content)
	}
	log.Printf("User: %s", username)
	log.Printf("Recent activities:\n %s", recentActivities)
	return recentActivities, nil
}
