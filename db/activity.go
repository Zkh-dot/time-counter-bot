package db

import (
	"errors"
	"log"
	"slices"
	"strings"

	"TimeCounterBot/common"
)

func addActivity(activity Activity) int64 {
	database := getPostgreSQLDatabase()

	insertActivitySQL := `INSERT INTO activities (user_id, name, parent_activity_id, is_leaf) 
		VALUES ($1, $2, $3, $4) RETURNING id
	`

	var currentActivityID int64
	err := database.QueryRow(
		insertActivitySQL, activity.UserID, activity.Name,
		activity.ParentActivityID, activity.IsLeaf,
	).Scan(&currentActivityID)

	if err != nil {
		log.Fatal(err)
	}

	return currentActivityID
}

func ParseAndAddActivity(userID common.UserID, activity string) {
	route := strings.Split(activity, " / ")
	var parentActivityID int64 = -1

	existingActivities := GetSimpleActivities(userID)

	for i, part := range route {
		isLeaf := false
		if i == len(route)-1 {
			isLeaf = true
		}

		idx := slices.IndexFunc(
			existingActivities,
			func(a Activity) bool {
				return a.Name == part && a.ParentActivityID == parentActivityID && a.IsLeaf == isLeaf
			},
		)
		if idx == -1 {
			parentActivityID = addActivity(
				Activity{
					UserID:           int64(userID),
					Name:             part,
					ParentActivityID: parentActivityID,
					IsLeaf:           isLeaf,
				},
			)
		} else {
			parentActivityID = existingActivities[idx].ID
		}
	}
}

func activityDFS(activities []Activity, vertex int, stack *[]string, ans *[]ActivityRoute) {
	if activities[vertex].IsLeaf {
		*ans = append(
			*ans,
			ActivityRoute{
				Name:   strings.Join(*stack, " / ") + " / " + activities[vertex].Name,
				LeafID: activities[vertex].ID,
			},
		)

		return
	}

	*stack = append(*stack, activities[vertex].Name)

	for childVertex := range activities {
		if childVertex == vertex {
			continue
		}

		if activities[childVertex].ParentActivityID == activities[vertex].ID {
			activityDFS(activities, childVertex, stack, ans)
		}
	}

	*stack = (*stack)[:len(*stack)-1]
}

func buildActivities(activities []Activity) []ActivityRoute {
	activitiesArray := make([]ActivityRoute, 0)

	for i, activity := range activities {
		if activity.ParentActivityID == -1 {
			stack := make([]string, 0)
			activityDFS(activities, i, &stack, &activitiesArray)
		}
	}

	return activitiesArray
}

func GetFullActivityNameByID(activityID int64, userID common.UserID) (string, error) {
	activities := GetFullActivities(userID)
	for _, activity := range activities {
		if activity.LeafID == activityID {
			return activity.Name, nil
		}
	}

	return "", errors.New("Activity not found")
}

func GetSimpleActivities(userID common.UserID) []Activity {
	database := getPostgreSQLDatabase()

	selectActivitySQL := `SELECT id, user_id, name, parent_activity_id, is_leaf FROM activities
		WHERE user_id = $1
	`

	rows, err := database.Query(selectActivitySQL, userID)
	if err != nil {
		log.Fatal(err)
	}
	if rows.Err() != nil {
		log.Fatal(rows.Err())
	}
	defer rows.Close()

	activities := make([]Activity, 0)

	for rows.Next() {
		activity := Activity{}

		err = rows.Scan(&activity.ID, &activity.UserID, &activity.Name,
			&activity.ParentActivityID, &activity.IsLeaf,
		)
		if err != nil {
			log.Fatal(err)
		}

		activities = append(activities, activity)
	}

	return activities
}

func GetFullActivities(userID common.UserID) []ActivityRoute {
	return buildActivities(GetSimpleActivities(userID))
}
