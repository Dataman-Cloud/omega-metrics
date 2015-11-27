package util

import (
	log "github.com/cihub/seelog"
)

func GetApps(clusterId string) ([]Application, error) {
	db := DB()
	applications := []Application{}
	rows, err := db.Query("select id, uid, cid, name, instances, status from application where cid = ?", clusterId)
	if err != nil {
		log.Error(err)
		return applications, err
	}

	for rows.Next() {
		var id int64
		var uid string
		var cid string
		var name string
		var instances int
		var status uint8
		if err = rows.Scan(&id, &uid, &cid, &name, &instances, &status); err != nil {
			log.Error(err)
			return applications, err
		}

		app := Application{
			Id:        id,
			Uid:       uid,
			Cid:       cid,
			Name:      name,
			Instances: instances,
			Status:    status,
		}
		applications = append(applications, app)
	}
	return applications, nil
}

func GetAllApps(uid string) ([]Application, error) {
	db := DB()
	applications := []Application{}
	rows, err := db.Query("select id, uid, cid, name, instances, status from application where uid = ?", uid)
	if err != nil {
		log.Error(err)
		return applications, err
	}

	for rows.Next() {
		var id int64
		var uid string
		var cid string
		var name string
		var instances int
		var status uint8
		if err = rows.Scan(&id, &uid, &cid, &name, &instances, &status); err != nil {
			log.Error(err)
			return applications, err
		}

		app := Application{
			Id:        id,
			Uid:       uid,
			Cid:       cid,
			Name:      name,
			Instances: instances,
			Status:    status,
		}
		applications = append(applications, app)
	}
	return applications, nil
}
