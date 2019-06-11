package sqlstore

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/grafana/grafana/pkg/models"
)

func TestUserDataAccess(t *testing.T) {

	Convey("Testing DB", t, func() {
		ss := InitTestDB(t)

		Convey("Creating a user", func() {
			cmd := &models.CreateUserCommand{
				Email: "usertest@test.com",
				Name:  "user name",
				Login: "user_test_login",
			}

			err := CreateUser(context.Background(), cmd)
			So(err, ShouldBeNil)

			Convey("Loading a user", func() {
				query := models.GetUserByIdQuery{Id: cmd.Result.Id}
				err := GetUserById(&query)
				So(err, ShouldBeNil)

				So(query.Result.Email, ShouldEqual, "usertest@test.com")
				So(query.Result.Password, ShouldEqual, "")
				So(query.Result.Rands, ShouldHaveLength, 10)
				So(query.Result.Salt, ShouldHaveLength, 10)
			})
		})

		Convey("Given 5 users", func() {
			var err error
			var cmd *models.CreateUserCommand
			users := []models.User{}
			for i := 0; i < 5; i++ {
				cmd = &models.CreateUserCommand{
					Email: fmt.Sprint("user", i, "@test.com"),
					Name:  fmt.Sprint("user", i),
					Login: fmt.Sprint("loginuser", i),
				}
				err = CreateUser(context.Background(), cmd)
				So(err, ShouldBeNil)
				users = append(users, cmd.Result)
			}

			Convey("Can return the first page of users and a total count", func() {
				query := models.SearchUsersQuery{Query: "", Page: 1, Limit: 3}
				err = SearchUsers(&query)

				So(err, ShouldBeNil)
				So(len(query.Result.Users), ShouldEqual, 3)
				So(query.Result.TotalCount, ShouldEqual, 5)
			})

			Convey("Can return the second page of users and a total count", func() {
				query := models.SearchUsersQuery{Query: "", Page: 2, Limit: 3}
				err = SearchUsers(&query)

				So(err, ShouldBeNil)
				So(len(query.Result.Users), ShouldEqual, 2)
				So(query.Result.TotalCount, ShouldEqual, 5)
			})

			Convey("Can return list of users matching query on user name", func() {
				query := models.SearchUsersQuery{Query: "use", Page: 1, Limit: 3}
				err = SearchUsers(&query)

				So(err, ShouldBeNil)
				So(len(query.Result.Users), ShouldEqual, 3)
				So(query.Result.TotalCount, ShouldEqual, 5)

				query = models.SearchUsersQuery{Query: "ser1", Page: 1, Limit: 3}
				err = SearchUsers(&query)

				So(err, ShouldBeNil)
				So(len(query.Result.Users), ShouldEqual, 1)
				So(query.Result.TotalCount, ShouldEqual, 1)

				query = models.SearchUsersQuery{Query: "USER1", Page: 1, Limit: 3}
				err = SearchUsers(&query)

				So(err, ShouldBeNil)
				So(len(query.Result.Users), ShouldEqual, 1)
				So(query.Result.TotalCount, ShouldEqual, 1)

				query = models.SearchUsersQuery{Query: "idontexist", Page: 1, Limit: 3}
				err = SearchUsers(&query)

				So(err, ShouldBeNil)
				So(len(query.Result.Users), ShouldEqual, 0)
				So(query.Result.TotalCount, ShouldEqual, 0)
			})

			Convey("Can return list of users matching query on email", func() {
				query := models.SearchUsersQuery{Query: "ser1@test.com", Page: 1, Limit: 3}
				err = SearchUsers(&query)

				So(err, ShouldBeNil)
				So(len(query.Result.Users), ShouldEqual, 1)
				So(query.Result.TotalCount, ShouldEqual, 1)
			})

			Convey("Can return list of users matching query on login name", func() {
				query := models.SearchUsersQuery{Query: "loginuser1", Page: 1, Limit: 3}
				err = SearchUsers(&query)

				So(err, ShouldBeNil)
				So(len(query.Result.Users), ShouldEqual, 1)
				So(query.Result.TotalCount, ShouldEqual, 1)
			})

			Convey("can return list users based on their auth type", func() {
				// add users to auth table
				for index, user := range users {
					authModule := "killa"

					// define every second user as ldap
					if index%2 == 0 {
						authModule = "ldap"
					}

					cmd2 := &models.SetAuthInfoCommand{
						UserId:     user.Id,
						AuthModule: authModule,
						AuthId:     "gorilla",
					}
					err = SetAuthInfo(cmd2)
					So(err, ShouldBeNil)
				}
				query := models.SearchUsersQuery{AuthModule: "ldap"}
				err = SearchUsers(&query)
				So(err, ShouldBeNil)

				So(query.Result.Users, ShouldHaveLength, 3)
				So(query.Result.Users[0].Name, ShouldEqual, "user0")
				So(query.Result.Users[1].Name, ShouldEqual, "user2")
				So(query.Result.Users[2].Name, ShouldEqual, "user4")
			})

			Convey("when a user is an org member and has been assigned permissions", func() {
				err = AddOrgUser(&models.AddOrgUserCommand{LoginOrEmail: users[1].Login, Role: models.ROLE_VIEWER, OrgId: users[0].OrgId, UserId: users[1].Id})
				So(err, ShouldBeNil)

				testHelperUpdateDashboardAcl(1, models.DashboardAcl{DashboardId: 1, OrgId: users[0].OrgId, UserId: users[1].Id, Permission: models.PERMISSION_EDIT})
				So(err, ShouldBeNil)

				err = SavePreferences(&models.SavePreferencesCommand{UserId: users[1].Id, OrgId: users[0].OrgId, HomeDashboardId: 1, Theme: "dark"})
				So(err, ShouldBeNil)

				Convey("when the user is deleted", func() {
					err = DeleteUser(&models.DeleteUserCommand{UserId: users[1].Id})
					So(err, ShouldBeNil)

					Convey("Should delete connected org users and permissions", func() {
						query := &models.GetOrgUsersQuery{OrgId: users[0].OrgId}
						err = GetOrgUsersForTest(query)
						So(err, ShouldBeNil)

						So(len(query.Result), ShouldEqual, 1)

						permQuery := &models.GetDashboardAclInfoListQuery{DashboardId: 1, OrgId: users[0].OrgId}
						err = GetDashboardAclInfoList(permQuery)
						So(err, ShouldBeNil)

						So(len(permQuery.Result), ShouldEqual, 0)

						prefsQuery := &models.GetPreferencesQuery{OrgId: users[0].OrgId, UserId: users[1].Id}
						err = GetPreferences(prefsQuery)
						So(err, ShouldBeNil)

						So(prefsQuery.Result.OrgId, ShouldEqual, 0)
						So(prefsQuery.Result.UserId, ShouldEqual, 0)
					})
				})

				Convey("when retreiving signed in user for orgId=0 result should return active org id", func() {
					ss.CacheService.Flush()

					query := &models.GetSignedInUserQuery{OrgId: users[1].OrgId, UserId: users[1].Id}
					err := ss.GetSignedInUserWithCache(query)
					So(err, ShouldBeNil)
					So(query.Result, ShouldNotBeNil)
					So(query.OrgId, ShouldEqual, users[1].OrgId)
					err = SetUsingOrg(&models.SetUsingOrgCommand{UserId: users[1].Id, OrgId: users[0].OrgId})
					So(err, ShouldBeNil)
					query = &models.GetSignedInUserQuery{OrgId: 0, UserId: users[1].Id}
					err = ss.GetSignedInUserWithCache(query)
					So(err, ShouldBeNil)
					So(query.Result, ShouldNotBeNil)
					So(query.Result.OrgId, ShouldEqual, users[0].OrgId)

					cacheKey := newSignedInUserCacheKey(query.Result.OrgId, query.UserId)
					_, found := ss.CacheService.Get(cacheKey)
					So(found, ShouldBeTrue)
				})
			})

			Convey("When batch disabling users", func() {
				userIdsToDisable := []int64{}
				for i := 0; i < 3; i++ {
					userIdsToDisable = append(userIdsToDisable, users[i].Id)
				}
				disableCmd := models.BatchDisableUsersCommand{UserIds: userIdsToDisable, IsDisabled: true}

				err = BatchDisableUsers(&disableCmd)
				So(err, ShouldBeNil)

				Convey("Should disable all provided users", func() {
					query := models.SearchUsersQuery{}
					err = SearchUsers(&query)

					So(query.Result.TotalCount, ShouldEqual, 5)
					for _, user := range query.Result.Users {
						shouldBeDisabled := false

						// Check if user id is in the userIdsToDisable list
						for _, disabledUserId := range userIdsToDisable {
							if user.Id == disabledUserId {
								So(user.IsDisabled, ShouldBeTrue)
								shouldBeDisabled = true
							}
						}

						// Otherwise user shouldn't be disabled
						if !shouldBeDisabled {
							So(user.IsDisabled, ShouldBeFalse)
						}
					}
				})
			})
		})

		Convey("Given one grafana admin user", func() {
			var err error
			createUserCmd := &models.CreateUserCommand{
				Email:   fmt.Sprint("admin", "@test.com"),
				Name:    fmt.Sprint("admin"),
				Login:   fmt.Sprint("admin"),
				IsAdmin: true,
			}
			err = CreateUser(context.Background(), createUserCmd)
			So(err, ShouldBeNil)

			Convey("Cannot make themselves a non-admin", func() {
				updateUserPermsCmd := models.UpdateUserPermissionsCommand{IsGrafanaAdmin: false, UserId: 1}
				updatePermsError := UpdateUserPermissions(&updateUserPermsCmd)

				So(updatePermsError, ShouldEqual, models.ErrLastGrafanaAdmin)

				query := models.GetUserByIdQuery{Id: createUserCmd.Result.Id}
				getUserError := GetUserById(&query)

				So(getUserError, ShouldBeNil)

				So(query.Result.IsAdmin, ShouldEqual, true)
			})
		})
	})
}

func GetOrgUsersForTest(query *models.GetOrgUsersQuery) error {
	query.Result = make([]*models.OrgUserDTO, 0)
	sess := x.Table("org_user")
	sess.Join("LEFT ", x.Dialect().Quote("user"), fmt.Sprintf("org_user.user_id=%s.id", x.Dialect().Quote("user")))
	sess.Where("org_user.org_id=?", query.OrgId)
	sess.Cols("org_user.org_id", "org_user.user_id", "user.email", "user.login", "org_user.role")

	err := sess.Find(&query.Result)
	return err
}
