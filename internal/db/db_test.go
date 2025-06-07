package db_test

import (
	"context"
	"database/sql"
	"fethcher/internal/db"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Test struct {
	ID       uint `gorm:"primaryKey"`
	Username string
}

var _ = Describe("Database", func() {
	var (
		mock   sqlmock.Sqlmock
		mockDb *sql.DB
		err    error
		testDB *db.PostgresDB
	)

	BeforeEach(func() {
		mockDb, mock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())

		dialector := postgres.New(postgres.Config{
			Conn:       mockDb,
			DriverName: "postgres",
		})

		gormDB, err := gorm.Open(dialector, &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())

		testDB = &db.PostgresDB{
			DB: gormDB,
		}

	})

	AfterEach(func() {
		mock.ExpectClose()
		Expect(mockDb.Close()).To(Succeed())
	})

	Describe("MigrateTable", func() {
		var err error

		BeforeEach(func() {
			// Reset the mock expectations before each test
			mock.ExpectQuery(`SELECT.*FROM information_schema\.tables.*`).
				WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(0))

			mock.ExpectExec(`^CREATE TABLE \"tests\".*$`).
				WillReturnResult(sqlmock.NewResult(0, 1))
		})
		JustBeforeEach(func() {
			err = testDB.MigrateTable(&Test{})
		})
		It("should migrate the table successfully", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(mock.ExpectationsWereMet()).To(Succeed())
		})
	})

	Describe("SaveToTable", func() {
		BeforeEach(func() {
			mock.ExpectQuery(`SELECT.*FROM "tests".*`).
				WillReturnRows(sqlmock.NewRows([]string{"id", "username"}))

			mock.ExpectBegin()

			mock.ExpectQuery(`^INSERT INTO "tests" \("username","id"\) VALUES \(\$1,\$2\),\(\$3,\$4\) RETURNING "id"$`).
				WithArgs("Alice", 1, "Bob", 2).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))

			mock.ExpectCommit()
		})

		It("should save records without errors", func() {
			err := testDB.SaveToTable(context.Background(), &[]Test{
				{ID: 1, Username: "Alice"},
				{ID: 2, Username: "Bob"},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(mock.ExpectationsWereMet()).To(Succeed())
		})
	})

	Describe("GetOneBy", func() {
		When("a record is found", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`SELECT \* FROM "tests" WHERE username = \$1 ORDER BY "tests"\."id" LIMIT \$2.*`).
					WithArgs("Alice", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).
						AddRow(1, "Alice"))
			})

			It("should return the correct record", func() {
				var result Test
				err := testDB.GetOneBy(context.Background(), "username", "Alice", &result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ID).To(Equal(uint(1)))
				Expect(result.Username).To(Equal("Alice"))
				Expect(mock.ExpectationsWereMet()).To(Succeed())
			})
		})

		When("no record is found", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`SELECT \* FROM "tests" WHERE username = \$1 ORDER BY "tests"\."id" LIMIT \$2.*`).
					WithArgs("Ghost", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			})

			It("should return ErrNotFound", func() {
				var result Test
				err := testDB.GetOneBy(context.Background(), "username", "Ghost", &result)
				Expect(err).To(Equal(db.ErrNotFound))
				Expect(mock.ExpectationsWereMet()).To(Succeed())
			})
		})
	})

	Describe("GetAllBy", func() {
		When("multiple records are found", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`SELECT \* FROM "tests" WHERE username IN \(\$1,\$2\).*`).
					WithArgs("Alice", "Bob").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).
						AddRow(1, "Alice").
						AddRow(2, "Bob"))
			})

			It("should return all matching records", func() {
				var results []Test
				err := testDB.GetAllBy(context.Background(), "username", []string{"Alice", "Bob"}, &results)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(2))
				Expect(results[0].Username).To(Equal("Alice"))
				Expect(results[1].Username).To(Equal("Bob"))
				Expect(mock.ExpectationsWereMet()).To(Succeed())
			})
		})

		When("an error occurs during query", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`SELECT \* FROM "tests" WHERE username.*`).
					WithArgs("Invalid").
					WillReturnError(sql.ErrConnDone)
			})

			It("should return an error", func() {
				var results []Test
				err := testDB.GetAllBy(context.Background(), "username", "Invalid", &results)
				Expect(err).To(MatchError(ContainSubstring("getting records by")))
				Expect(mock.ExpectationsWereMet()).To(Succeed())
			})
		})
	})

})
