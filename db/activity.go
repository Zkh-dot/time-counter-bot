package db

import (
	"errors"
	"log"
	"slices"
	"strings"

	"TimeCounterBot/common"
)

func addActivity(activity Activity) int64 {
	// mutex.Lock()
	// defer mutex.Unlock()

	// database, err := sql.Open("sqlite3", "database.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer database.Close()

	database := getPostgreSQLDatabase()

	insertActivitySQL := `INSERT INTO activities (user_id, name, parent_activity_id, is_leaf) 
		VALUES (?, ?, ?, ?)
	`
	row, err := database.Exec(
		insertActivitySQL, activity.UserID, activity.Name,
		activity.ParentActivityID, activity.IsLeaf,
	)
	if err != nil {
		log.Fatal(err)
	}

	currentActivityID, err := row.LastInsertId()
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

func ParseAndAddActivityDeprecated(userID common.UserID, activity string) {
	// mutex.Lock()
	// defer mutex.Unlock()

	// database, err := sql.Open("sqlite3", "database.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer database.Close()

	database := getPostgreSQLDatabase()

	insertActivitySQL := `INSERT INTO activities (user_id, name, parent_activity_id, is_leaf) 
		VALUES (?, ?, ?, ?)
	`
	selectActivitySQL := `SELECT id FROM activities
		WHERE user_id == ? AND name == ? AND parent_activity_id == ? AND is_leaf == ?
	`

	route := strings.Split(activity, " / ")

	var parentActivityID int64 = -1

	for i, part := range route {
		isLeaf := false
		if i == len(route)-1 {
			isLeaf = true
		}

		rows, err := database.Query(selectActivitySQL, userID, part, parentActivityID, isLeaf)
		if err != nil {
			log.Fatal(err)
		}
		if rows.Err() != nil {
			log.Fatal(rows.Err())
		}

		defer rows.Close()

		found := false

		for rows.Next() {
			err = rows.Scan(&parentActivityID)
			if err != nil {
				log.Fatal(err)
			}

			found = true
		}

		if !found {
			row, err := database.Exec(insertActivitySQL, userID, part, parentActivityID, isLeaf)
			if err != nil {
				log.Fatal(err)
			}

			parentActivityID, err = row.LastInsertId()
			if err != nil {
				log.Fatal(err)
			}
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
	// mutex.Lock()
	// defer mutex.Unlock()

	// database, err := sql.Open("sqlite3", "database.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer database.Close()

	database := getPostgreSQLDatabase()

	selectActivitySQL := `SELECT id, user_id, name, parent_activity_id, is_leaf FROM activities
		WHERE user_id == ?
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
