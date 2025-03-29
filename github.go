package ghsummary

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Activity struct {
	Type       string
	Repository string
	Content    string
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

func GetUserActivity(username string, maxEvents int) (string, error) {
	maxEvents = min(maxEvents, 100) // Limit to 100 events. In the future I may want to add pagination.
	log.Printf("Fetching activity for user: %s with max events: %d", username, maxEvents)
	url := fmt.Sprintf("https://api.github.com/users/%s/events?per_page=%d", username, maxEvents)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error making HTTP request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Non-OK HTTP status: %s", resp.Status)
		return "", fmt.Errorf("failed to fetch activity: %s", resp.Status)
	}

	var events []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		log.Printf("Error decoding JSON response: %v", err)
		return "", err
	}

	log.Printf("Successfully fetched %d events", len(events))

	// Simplify activity data for LLM
	activities := []Activity{}
	for _, event := range events {
		if len(activities) >= maxEvents {
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

				activities = append(activities, Activity{
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
				commits, ok := payload["commits"].([]interface{})
				if !ok {
					log.Printf("[%s] Error parsing commits data", id)
					continue
				}
				messages := ""
				for _, commit := range commits {
					commitData, ok := commit.(map[string]interface{})
					if !ok {
						log.Printf("[%s] Error parsing commit data", id)
						continue
					}
					message, ok := commitData["message"].(string)
					if !ok {
						log.Printf("[%s] Error parsing commit message", id)
						continue
					}
					messages += message + "\n"
				}

				if messages == "" {
					log.Printf("[%s] No commit messages found", id)
					continue
				}

				activities = append(activities, Activity{
					Type:       eventType,
					Repository: repo,
					Content:    messages,
				})

			default:
				log.Printf("[%s] Unsupported event type: %s", id, eventType)
			}
		}
	}
	recentActivities := fmt.Sprintf("Recent activities for user %s:\n", username)
	recentActivities += "User's pronounces: he/him\n"
	for _, activity := range activities {
		recentActivities += fmt.Sprintf("Type: %s\nRepository: %s\nContent: %s\n\n", activity.Type, activity.Repository, activity.Content)
	}

	log.Printf("User: %s", username)
	log.Printf("Recent activities:\n %s", recentActivities)
	return recentActivities, nil
}
