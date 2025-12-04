package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func AddTask(task *Task) (int64, error) {
	if DB == nil {
		return 0, sql.ErrConnDone
	}
	query := `INSERT INTO scheduler(date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	task.ID = strconv.FormatInt(id, 10)
	return id, nil
}

func parseSearchDate(s string) (string, bool) {
	t, err := time.Parse("02.01.2006", s)
	if err != nil {
		return "", false
	}
	return t.Format("20060102"), true
}

func Tasks(limit int, search string) ([]*Task, error) {

	if DB == nil {
		return nil, sql.ErrConnDone
	}

	var rows *sql.Rows
	var err error

	if search != "" {
		if d, ok := parseSearchDate(search); ok {
			query := `SELECT id, date, title, comment, repeat
			          FROM scheduler
			          WHERE date = ?
			          ORDER BY date
			          LIMIT ?`
			rows, err = DB.Query(query, d, limit)
			if err != nil {
				return nil, err
			}
			return scanTasks(rows)
		}
	}

	if search != "" {
		like := "%" + search + "%"
		query := `SELECT id, date, title, comment, repeat
		          FROM scheduler
		          WHERE title LIKE ? OR comment LIKE ?
		          ORDER BY date
		          LIMIT ?`
		rows, err = DB.Query(query, like, like, limit)
		if err != nil {
			return nil, err
		}
		return scanTasks(rows)
	}

	query := `SELECT id, date, title, comment, repeat
	          FROM scheduler
	          ORDER BY date
	          LIMIT ?`

	rows, err = DB.Query(query, limit)
	if err != nil {
		return nil, err
	}

	return scanTasks(rows)
}

func scanTasks(rows *sql.Rows) ([]*Task, error) {
	defer rows.Close()

	tasks := []*Task{}

	for rows.Next() {
		var (
			id      int64
			date    string
			title   string
			comment string
			repeat  string
		)

		if err := rows.Scan(&id, &date, &title, &comment, &repeat); err != nil {
			return nil, err
		}

		tasks = append(tasks, &Task{
			ID:      strconv.FormatInt(id, 10),
			Date:    date,
			Title:   title,
			Comment: comment,
			Repeat:  repeat,
		})
	}

	if tasks == nil {
		tasks = []*Task{}
	}

	return tasks, nil
}

func GetTask(id string) (*Task, error) {
	row := DB.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id)

	var t Task
	err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}
	return &t, nil
}

func UpdateTask(t *Task) error {
	if DB == nil {
		return sql.ErrConnDone
	}

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := DB.Exec(query, t.Date, t.Title, t.Comment, t.Repeat, t.ID)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("incorrect id for updating task")
	}

	return nil
}

func DeleteTask(id string) error {
	if DB == nil {
		return sql.ErrConnDone
	}
	query := `DELETE FROM scheduler WHERE id = ?`
	res, err := DB.Exec(query, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

func UpdateDate(next string, id string) error {
	if DB == nil {
		return sql.ErrConnDone
	}
	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := DB.Exec(query, next, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}
