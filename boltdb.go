package main 

import (
    "github.com/boltdb/bolt"
    "strconv"
    "encoding/json"
    "fmt"
)

type BoltDB struct {
    handler *bolt.DB
}

func BoltInit() (db *BoltDB,err error) {
    db=&BoltDB{}
    myLogf(LogInfo,"Initializing Bolt Database at %s\n",settings.DBPath)
    db.handler,err=bolt.Open(settings.DBPath,0600,nil)
    if err!=nil {
        return
    }
    //пытаемся создать bucket
    err=db.handler.Update(func(tx *bolt.Tx) (err error){
        _,err=tx.CreateBucketIfNotExists([]byte(settings.DBTable))
        return
    })
    return
}

func (db *BoltDB) AddStation(code int,node string,user int) (err error) {
    myLogf(LogDebug,"BoltDB: Trying add station %s in node %s for user %d\n",stationByCode(code),node,user)
    err=db.handler.Update(func(tx *bolt.Tx) error{
        bck:=tx.Bucket([]byte(settings.DBTable))
        userinfobytes:=bck.Get([]byte(strconv.Itoa(user)))
        userinfo:=DBUserInfo{}
        if len(userinfobytes)!=0 {
            err:=json.Unmarshal(userinfobytes,&userinfo)
            if err!=nil {
                return err
            }
        } else {
            userinfo.ID=user
        }
        userinfo.Stations=append(userinfo.Stations,NSPair{Node:node,Station:code})
        buf,err:=json.Marshal(userinfo)
        if err!=nil {
            return err
        }
        err=bck.Put([]byte(strconv.Itoa(user)),buf)
        return err
    })
    return
}

func (db *BoltDB) DelStation(code int,user int) (err error) {
    myLogf(LogDebug,"BoltDB: Trying add station %s for user %d\n",stationByCode(code),user)
    err=db.handler.Update(func(tx *bolt.Tx) error{
        bck:=tx.Bucket([]byte(settings.DBTable))
        userinfobytes:=bck.Get([]byte(strconv.Itoa(user)))
        userinfo:=DBUserInfo{}
        if len(userinfobytes)==0 {
            return nil
        }
        err:=json.Unmarshal(userinfobytes,&userinfo)
        if err!=nil {
            return err
        }
        i:=0
        for ;i<len(userinfo.Stations);i++ {
            if userinfo.Stations[i].Station==code {
                break
            }
        }
        if i>=len(userinfo.Stations) {
            return fmt.Errorf("User`s (%d) record not contains code %d",user,code)
        }
        //удаление
        userinfo.Stations=append(userinfo.Stations[:i],userinfo.Stations[i+1:]...)
        buf,err:=json.Marshal(userinfo)
        if err!=nil {
            return err
        }
        err=bck.Put([]byte(strconv.Itoa(user)),buf)
        return err
    })
    return
}

func (db *BoltDB) GetUserStations(user int) (ret DBUserInfo,err error) {
    myLogf(LogDebug,"BoltDB: Extracting data for user %d\n",user)
    err=db.handler.View(func(tx *bolt.Tx) error{
        bck:=tx.Bucket([]byte(settings.DBTable))
        userinfobytes:=bck.Get([]byte(strconv.Itoa(user)))
        if len(userinfobytes)==0 {
            return nil
        }
        err:=json.Unmarshal(userinfobytes,&ret)
        return err
    })
    return
}

func (db *BoltDB) Close() error {
    return db.handler.Close()
}