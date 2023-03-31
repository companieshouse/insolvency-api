package dao

import (
	"testing"

	"github.com/companieshouse/insolvency-api/config"
	"github.com/companieshouse/insolvency-api/models"

	"github.com/stretchr/testify/assert"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func setDriverUp() (MongoService, mtest.CommandError, models.InsolvencyResourceDao, *mtest.Options, []models.PractitionerResourceDao) {
	client = &mongo.Client{}
	cfg, _ := config.Get()
	dataBase := NewGetMongoDatabase("mongoDBURL", "databaseName")

	insolvencyResourcePractitionersDao := models.InsolvencyResourcePractitionersDao{
		"VM04221441": "/transactions/168570-809316-704268/insolvency/practitioners/VM04221441",
	}

	mongoService := MongoService{
		db:             dataBase,
		CollectionName: cfg.MongoCollection,
	}

	commandError := mtest.CommandError{
		Code:    1,
		Message: "Message",
		Name:    "Name",
		Labels:  []string{"label1"},
	}

	practitionerResourceDao := models.PractitionerResourceDao{}
	practitionerResourceDao.Data.IPCode = "IPCode"
	practitionerResourceDao.Data.FirstName = "FirstName"
	practitionerResourceDao.Data.LastName = "LastName"
	practitionerResourceDao.Data.TelephoneNumber = "TelephoneNumber"
	practitionerResourceDao.Data.Email = "Email"
	practitionerResourceDao.Data.Address = models.AddressResourceDao{}
	practitionerResourceDao.Data.Role = "Role"
	practitionerResourceDao.Data.Links = models.PractitionerResourceLinksDao{}
	practitionerResourceDao.Data.Appointment = &models.AppointmentResourceDao{}

	practitioners := []models.PractitionerResourceDao{}

	dataInsolvency := models.InsolvencyResourceDao{}
	dataInsolvency.Data.CompanyNumber = "CompanyNumber"
	dataInsolvency.Data.CaseType = "CaseType"
	dataInsolvency.Data.CompanyName = "CompanyName"
	dataInsolvency.Data.Practitioners = &insolvencyResourcePractitionersDao
	dataInsolvency.Data.Attachments = []models.AttachmentResourceDao{}
	dataInsolvency.Data.Resolution = &models.ResolutionResourceDao{}
	dataInsolvency.Data.StatementOfAffairs = &models.StatementOfAffairsResourceDao{}
	dataInsolvency.Data.Links = models.InsolvencyResourceLinksDao{}

	expectedInsolvency := models.InsolvencyResourceDao{
		ID:            primitive.NewObjectID(),
		TransactionID: "TransactionID",
	}

	expectedInsolvency.Data = dataInsolvency.Data

	opts := mtest.NewOptions().DatabaseName("databaseName").ClientType(mtest.Mock)

	return mongoService, commandError, expectedInsolvency, opts, append(practitioners, practitionerResourceDao)
}

func TestUnitUpdateAttachmentStatusDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)

	defer mt.Close()

	bsonDataAttachment := bson.M{
		"id":     "ID",
		"type":   "type",
		"status": "status",
	}

	bsonAttachmentArrays := bson.A{bsonDataAttachment}

	bsonPractitionerLinksMap := bson.M{
		"PractionerID1": "PractitionerLink1",
		"PractionerID2": "PractitionerLink2",
	}

	bsonInsolvency := bson.D{
		{"company_number", "CompanyNumber"},
		{"case_type", "CaseType"},
		{"company_name", "CompanyName"},
		{"practitioners", bsonPractitionerLinksMap},
		{"attachments", bsonAttachmentArrays},
	}

	mt.Run("UpdateAttachmentStatus runs successfully", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
			bson.E{Key: "upserted", Value: 1},
		))

		mongoService.db = mt.DB
		code, err := mongoService.UpdateAttachmentStatus("transactionID", "attachmentID", "avStatus")

		assert.Nil(t, err)
		assert.Equal(t, code, 204)
	})

	mt.Run("UpdateAttachmentStatus runs with error on FindOne", func(mt *mtest.T) {

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB

		_, err := mongoService.UpdateAttachmentStatus("transactionID", "attachmentID", "avStatus")

		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
	})

	mt.Run("UpdateAttachmentStatus runs successfully with status not processed", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
			bson.E{Key: "upserted", Value: 1},
		))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		code, err := mongoService.UpdateAttachmentStatus("transactionID", "attachmentID", "avStatus")

		assert.NotNil(t, err)
		assert.Equal(t, code, 500)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID] - could not update status of attachment with id [attachmentID]")
	})

	mt.Run("UpdateAttachmentStatus runs with error on UpdateOne", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mongoService.db = mt.DB
		code, err := mongoService.UpdateAttachmentStatus("transactionID", "attachmentID", "avStatus")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID] - could not update status of attachment with id [attachmentID]")
		assert.Equal(t, code, 500)

	})

	mt.Run("UpdateAttachmentStatus runs successfully with ModifiedCount zero", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 0},
			bson.E{Key: "upserted", Value: 1},
		))

		mongoService.db = mt.DB
		code, err := mongoService.UpdateAttachmentStatus("transactionID", "attachmentID", "avStatus")

		assert.NotNil(t, err)
		assert.Equal(t, code, 404)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID] - attachment with id [attachmentID] not found")
	})
}

func TestUnitCreateInsolvencyResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("CreateInsolvencyResource with error findone", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		_, err := mongoService.CreateInsolvencyResource(&expectedInsolvency)

		assert.NotNil(t, err.Error())
		assert.Equal(t, err.Error(), "there was a problem creating an insolvency case for this transaction id: (Name) Message")
	})

	mt.Run("CreateInsolvencyResource with successful created one", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
		}))

		mongoService.db = mt.DB
		_, err := mongoService.CreateInsolvencyResource(&expectedInsolvency)

		assert.Equal(t, err.Error(), "an insolvency case already exists for this transaction id")
	})
}

func TestUnitGetInsolvencyPractitionersResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("GetInsolvencyPractitionersResource runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		_, _, err := mongoService.GetInsolvencyPractitionersResource("transactionID")

		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction transactionID")
	})

	mt.Run("GetInsolvencyPractitionersResource runs successfully", func(mt *mtest.T) {
		id1 := primitive.NewObjectID()
		id2 := primitive.NewObjectID()

		first := mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", id1},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", expectedInsolvency.Data},
		})

		second := mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.NextBatch, bson.D{
			{"_id", id2},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", expectedInsolvency.Data},
		})

		killCursors := mtest.CreateCursorResponse(0, "models.InsolvencyResourceDao", mtest.NextBatch)
		mt.AddMockResponses(first, second, killCursors)

		mongoService.db = mt.DB
		insolvencyResource, _, err := mongoService.GetInsolvencyPractitionersResource("transactionID")

		assert.Nil(t, err)
		assert.NotNil(t, insolvencyResource)
		assert.Equal(t, insolvencyResource.Data.Etag, expectedInsolvency.Data.Etag)
		assert.Equal(t, insolvencyResource.Data.Kind, expectedInsolvency.Data.Kind)
	})
}

func TestUnitCreatePractitionerResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, _, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	practitionerResourceDao := models.PractitionerResourceDao{}
	practitionerResourceDao.Data.PractitionerId = "practitionerID"
	practitionerResourceDao.Data.IPCode = "IPCode"

	mt.Run("CreatePractitionerResource runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		code, err := mongoService.CreatePractitionerResource(&practitionerResourceDao, "transactionID")

		assert.Equal(t, code, 500)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction transactionID (insert practitioner to collection)")
	})

	mt.Run("CreatePractitionerResource runs with error on duplicate key insert", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   1,
			Code:    11000,
			Message: "duplicate key error",
		}))

		mongoService.db = mt.DB
		code, err := mongoService.CreatePractitionerResource(&practitionerResourceDao, "transactionID")

		assert.Equal(t, code, 500)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction transactionID (insert practitioner to collection)")
	})

	mt.Run("CreatePractitionerResource runs successfully with a Practitioner", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		practitionerResourceDao = models.PractitionerResourceDao{}
		practitionerResourceDao.Data.PractitionerId = "practitionerID"
		practitionerResourceDao.Data.IPCode = "IPCode"

		mongoService.db = mt.DB
		code, err := mongoService.CreatePractitionerResource(&practitionerResourceDao, "transactionID")

		assert.Nil(t, err)
		assert.Equal(t, code, 201)
	})
}

func TestUnitGetPractitionerAppointmentDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, _, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("GetPractitionerAppointment runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		practitioner, err := mongoService.GetPractitionerAppointment("practitionerID", "transactionID")

		assert.Nil(t, practitioner)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
	})

	mt.Run("GetPractitionerAppointment failed on findone", func(mt *mtest.T) {

		bsonData := bson.M{
			"appointed_on": "appointedon",
			"made_by":      "madeby",
			"links":        "appointmentResourceLinksDao",
			"last_name":    "LastName",
			"etag":         "etag",
			"kind":         "kind",
		}

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		first := mtest.CreateCursorResponse(1, "models.AppointmentResourceDao", mtest.FirstBatch, bson.D{
			{"data", bsonData},
		})

		mt.AddMockResponses(first)

		mongoService.db = mt.DB
		appointmentResource, err := mongoService.GetPractitionerAppointment("practitionerID", "transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
		assert.Nil(t, appointmentResource)
	})

	mt.Run("GetPractitionerAppointment failed on decoding model", func(mt *mtest.T) {

		bsonData := bson.M{
			"appointed_on": "appointedon",
			"made_by":      "madeby",
			"links":        "appointmentResourceLinksDao",
			"last_name":    "LastName",
			"etag":         "etag",
			"kind":         "kind",
		}

		first := mtest.CreateCursorResponse(1, "models.AppointmentResourceDao", mtest.FirstBatch, bson.D{
			{"data", bsonData},
		})

		mt.AddMockResponses(first)

		mongoService.db = mt.DB
		appointmentResource, err := mongoService.GetPractitionerAppointment("practitionerID", "transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
		assert.Nil(t, appointmentResource)
	})

	mt.Run("GetPractitionerAppointment runs successfully", func(mt *mtest.T) {
		appointmentResourceLinksDao := models.AppointmentResourceLinksDao{}

		bsonData := bson.M{
			"appointed_on": "appointedon",
			"made_by":      "madeby",
			"links":        appointmentResourceLinksDao,
			"last_name":    "LastName",
			"etag":         "etag",
			"kind":         "kind",
		}

		first := mtest.CreateCursorResponse(1, "models.AppointmentResourceDao", mtest.FirstBatch, bson.D{
			{"data", bsonData},
		})

		mt.AddMockResponses(first)

		mongoService.db = mt.DB
		insolvencyResource, err := mongoService.GetPractitionerAppointment("practitionerID", "transactionID")

		assert.Nil(t, err)
		assert.NotNil(t, insolvencyResource)
	})
}
func TestUnitGetPractitionersResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, _, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("GetPractitionersResource runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		practitioner, err := mongoService.GetPractitionersResource([]string{"practitionerID"})

		assert.Nil(t, practitioner)
		assert.Equal(t, err.Error(), "no practitioner found for practitioner id(s) [practitionerID]")
	})

	mt.Run("GetPractitionersResource runs successfully", func(mt *mtest.T) {
		bsonData := bson.M{
			"id":               "ID",
			"ip_code":          "IPCode",
			"first_name":       "FirstName",
			"last_name":        "LastName",
			"telephone_number": "TelephoneNumber",
			"email":            "Email",
		}

		first := mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"data", bsonData},
		})
		second := mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.NextBatch, bson.D{
			{"data", bsonData},
		})

		killCursors := mtest.CreateCursorResponse(0, "models.InsolvencyResourceDao", mtest.NextBatch)
		mt.AddMockResponses(first, second, killCursors)

		mongoService.db = mt.DB
		insolvencyResource, err := mongoService.GetPractitionersResource([]string{"practitionerID"})

		assert.Nil(t, err)
		assert.NotNil(t, insolvencyResource)
	})
}

func TestUnitDeletePractitionerDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	bsonPractitionerLinksMap := bson.M{
		"VM04221441":    "PractitionerLink1",
		"PractionerID2": "PractitionerLink2",
	}

	bsonInsolvency := bson.D{
		{"company_number", "CompanyNumber"},
		{"case_type", "CaseType"},
		{"company_name", "CompanyName"},
		{"practitioners", bsonPractitionerLinksMap},
	}

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("DeletePractitioner runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		_, err := mongoService.DeletePractitioner("practitionerID", "transactionID")

		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id transactionID")
	})

	mt.Run("DeletePractitioner runs with error with missing practitioner links", func(mt *mtest.T) {

		bsonInsolvency := bson.D{
			{"company_number", "CompanyNumber"},
			{"case_type", "CaseType"},
			{"company_name", "CompanyName"},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mongoService.db = mt.DB
		code, err := mongoService.DeletePractitioner("practitionerID", "transactionID")

		assert.Equal(t, code, 400)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id transactionID no insolvency practitioners found")
	})

	mt.Run("DeletePractitioner runs with error when practitioner ID not matched", func(mt *mtest.T) {

		bsonInsolvency := bson.D{
			{"company_number", "CompanyNumber"},
			{"case_type", "CaseType"},
			{"company_name", "CompanyName"},
			{"practitioners", bsonPractitionerLinksMap},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mongoService.db = mt.DB
		code, err := mongoService.DeletePractitioner("practitionerID", "transactionID")

		assert.Equal(t, 404, code)
		assert.Equal(t, "there was a problem handling your request for transaction id transactionID not able to find practitioner practitionerID to delete", err.Error())
	})

	mt.Run("DeletePractitioner run successfully with correct practitionerID", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(bson.D{{"ok", 1}, {"acknowledged", true}, {"n", 1}})

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"value", bson.D{
				{"_id", expectedInsolvency.ID},
				{"transaction_id", expectedInsolvency.TransactionID},
				{"etag", expectedInsolvency.Data.Etag},
				{"kind", expectedInsolvency.Data.Kind},
				{"data", bsonInsolvency},
			}},
		})

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"nModified", 1},
		})

		mongoService.db = mt.DB
		code, err := mongoService.DeletePractitioner("VM04221441", "transactionID")

		assert.Nil(t, err)
		assert.Equal(t, code, 204)

	})

	mt.Run("DeletePractitioner run successfully with ModifiedCount", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(bson.D{{"ok", 1}, {"acknowledged", true}, {"n", 1}})

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"value", bson.D{
				{"_id", expectedInsolvency.ID},
				{"transaction_id", expectedInsolvency.TransactionID},
				{"etag", expectedInsolvency.Data.Etag},
				{"kind", expectedInsolvency.Data.Kind},
				{"data", bsonInsolvency},
			}},
		})

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"nModified", 1},
		})

		mongoService.db = mt.DB
		code, err := mongoService.DeletePractitioner("VM04221441", "transactionID")

		assert.Nil(t, err)
		assert.Equal(t, code, 204)

	})

	mt.Run("DeletePractitioner with ModifiedCount zero", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(bson.D{{"ok", 1}, {"acknowledged", true}, {"n", 1}})

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"value", bson.D{
				{"_id", expectedInsolvency.ID},
				{"transaction_id", expectedInsolvency.TransactionID},
				{"etag", expectedInsolvency.Data.Etag},
				{"kind", expectedInsolvency.Data.Kind},
				{"data", bsonInsolvency},
			}},
		})

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"nModified", 0},
		})

		mongoService.db = mt.DB
		code, err := mongoService.DeletePractitioner("VM04221441", "transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id transactionID - not able to update insolvency practitioners VM04221441")
		assert.Equal(t, code, 404)

	})

	mt.Run("DeletePractitioner runs failed to delete practitioner after sucessfully deleted appointment", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"nModified", 1},
		})

		mongoService.db = mt.DB
		code, err := mongoService.DeletePractitioner("VM04221441", "transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id transactionID not able to delete practitioners")
		assert.Equal(t, code, 500)

	})

	mt.Run("DeletePractitioner runs failed to delete appointment", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(bson.D{{"ok", 0}, {"acknowledged", true}, {"n", 0}})

		mongoService.db = mt.DB
		code, err := mongoService.DeletePractitioner("VM04221441", "transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id transactionID not able to delete practitioners appointment")
		assert.Equal(t, code, 500)

	})
}

func TestUnitAppointPractitionerDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, _, opts, _ := setDriverUp()

	appointmentResourceDao := models.AppointmentResourceDao{}
	appointmentResourceDao.Data.AppointedOn = "AppointedOn"
	appointmentResourceDao.Data.MadeBy = "MadeBy"
	appointmentResourceDao.Data.Links = models.AppointmentResourceLinksDao{}
	appointmentResourceDao.PractitionerId = "PractitionerID"

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("UpdatePractitionerAppointment runs with error on updateCollection after insert into appointment collection", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		statusCode, err := mongoService.UpdatePractitionerAppointment(&appointmentResourceDao, "transactionID", "practitionerID")

		assert.Equal(t, statusCode, 500)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id transactionID - not able to update practitioner's appointment practitionerID")
	})
}

func TestUnitDeletePractitionerAppointmentDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	bsonPractitionerLinksMap := bson.M{
		"PractionerID1": "PractitionerLink1",
		"PractionerID2": "PractitionerLink2",
	}

	bsonInsolvency := bson.D{
		{"company_number", "CompanyNumber"},
		{"case_type", "CaseType"},
		{"company_name", "CompanyName"},
		{"practitioners", bsonPractitionerLinksMap},
	}

	mt.Run("DeletePractitionerAppointment runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB

		_, err := mongoService.DeletePractitionerAppointment("transactionID", "practitionerID")

		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id transactionID")
	})

	mt.Run("DeletePractitionerAppointment runs with error with missing appointment link", func(mt *mtest.T) {

		practitionerLinks := models.PractitionerResourceLinksDao{
			Self: "",
		}

		bsonInsolvency := bson.D{
			{"company_number", "CompanyNumber"},
			{"case_type", "CaseType"},
			{"company_name", "CompanyName"},
			{"practitioners", bsonPractitionerLinksMap},
			{"links", practitionerLinks},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mongoService.db = mt.DB
		code, err := mongoService.DeletePractitionerAppointment("transactionID", "practitionerID")

		assert.Equal(t, code, 404)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id transactionID - no practitioner's appointment found")
	})

	mt.Run("DeletePractitionerAppointment run successfully with ModifiedCount", func(mt *mtest.T) {

		practitionerLinks := models.PractitionerResourceLinksDao{
			Self:        "",
			Appointment: "/transactions/168570-809316-704268/insolvency/practitioners/VM04221441/appointment",
		}

		bsonInsolvency := bson.D{
			{"company_number", "CompanyNumber"},
			{"case_type", "CaseType"},
			{"company_name", "CompanyName"},
			{"practitioners", bsonPractitionerLinksMap},
			{"links", practitionerLinks},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"value", bson.D{
				{"_id", expectedInsolvency.ID},
				{"transaction_id", expectedInsolvency.TransactionID},
				{"etag", expectedInsolvency.Data.Etag},
				{"kind", expectedInsolvency.Data.Kind},
				{"data", bsonInsolvency},
			}},
		})

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"nModified", 1},
		})

		mongoService.db = mt.DB
		code, err := mongoService.DeletePractitionerAppointment("168570-809316-704268", "VM04221441")

		assert.Nil(t, err)
		assert.Equal(t, code, 204)

	})

	mt.Run("DeletePractitionerAppointment with ModifiedCount zero", func(mt *mtest.T) {
		practitionerLinks := models.PractitionerResourceLinksDao{
			Self:        "",
			Appointment: "/transactions/168570-809316-704268/insolvency/practitioners/VM04221441/appointment",
		}

		bsonInsolvency := bson.D{
			{"company_number", "CompanyNumber"},
			{"case_type", "CaseType"},
			{"company_name", "CompanyName"},
			{"practitioners", bsonPractitionerLinksMap},
			{"links", practitionerLinks},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(bson.D{{"ok", 1}, {"acknowledged", true}, {"n", 1}})

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"value", bson.D{
				{"_id", expectedInsolvency.ID},
				{"transaction_id", expectedInsolvency.TransactionID},
				{"etag", expectedInsolvency.Data.Etag},
				{"kind", expectedInsolvency.Data.Kind},
				{"data", bsonInsolvency},
			}},
		})

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"nModified", 0},
		})

		mongoService.db = mt.DB
		code, err := mongoService.DeletePractitionerAppointment("168570-809316-704268", "VM04221441")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id 168570-809316-704268 - not able to update insolvency practitioners VM04221441")
		assert.Equal(t, code, 404)

	})

	mt.Run("DeletePractitionerAppointment runs failed to delete practitioner after sucessfully deleted appointment", func(mt *mtest.T) {
		practitionerLinks := models.PractitionerResourceLinksDao{
			Self:        "",
			Appointment: "/transactions/168570-809316-704268/insolvency/practitioners/VM04221441/appointment",
		}

		bsonInsolvency := bson.D{
			{"company_number", "CompanyNumber"},
			{"case_type", "CaseType"},
			{"company_name", "CompanyName"},
			{"practitioners", bsonPractitionerLinksMap},
			{"links", practitionerLinks},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(bson.D{{"ok", 1}, {"acknowledged", true}, {"n", 0}})

		mongoService.db = mt.DB
		code, err := mongoService.DeletePractitionerAppointment("168570-809316-704268", "VM04221441")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id 168570-809316-704268 - not able to delete practitioners appointment")
		assert.Equal(t, code, 500)

	})

	mt.Run("DeletePractitionerAppointment runs and failed to fetch insolvency resource", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
			bson.E{Key: "upserted", Value: 1},
		))

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		code, err := mongoService.DeletePractitionerAppointment("transactionID", "practitionerID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id transactionID")
		assert.Equal(t, code, 500)

	})
}

func TestUnitAddAttachmentToInsolvencyResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("AddAttachmentToInsolvencyResource runs successfully with findone", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
			bson.E{Key: "upserted", Value: 1},
		))

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", expectedInsolvency.Data.Practitioners},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		attachmentDao, err := mongoService.AddAttachmentToInsolvencyResource("transactionID", "fileID", "attachmentType")

		assert.Nil(t, err)
		assert.NotNil(t, attachmentDao)
		assert.Equal(t, attachmentDao.ID, "fileID")
		assert.Equal(t, attachmentDao.Type, "attachmentType")
		assert.Equal(t, attachmentDao.Status, "submitted")

	})

	mt.Run("AddAttachmentToInsolvencyResource runs with MatchedCount OR ModifiedCount zero", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 0},
			bson.E{Key: "nModified", Value: 0},
			bson.E{Key: "upserted", Value: 1},
		))

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", expectedInsolvency.Data.Practitioners},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		attachmentDao, err := mongoService.AddAttachmentToInsolvencyResource("transactionID", "fileID", "attachmentType")

		assert.NotNil(t, err)
		assert.Nil(t, attachmentDao)
		assert.Equal(t, err.Error(), "no documents updated")
	})

	mt.Run("AddAttachmentToInsolvencyResource runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		_, err := mongoService.AddAttachmentToInsolvencyResource("transactionID", "fileID", "attachmentType")

		assert.Equal(t, err.Error(), "error updating mongo for transaction [transactionID]: [(Name) Message]")
	})
}

func TestUnitGetAttachmentResourcesDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	bsonPractitionerLinksMap := bson.M{
		"PractionerID1": "PractitionerLink1",
		"PractionerID2": "PractitionerLink2",
	}

	bsonDataAttachment := bson.M{
		"id":     "ID",
		"type":   "type",
		"status": "status",
	}

	bsonAttachmentArrays := bson.A{bsonDataAttachment}

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("GetAttachmentResources runs successfully", func(mt *mtest.T) {
		bsonInsolvency := bson.D{
			{"company_number", "CompanyNumber"},
			{"case_type", "CaseType"},
			{"company_name", "CompanyName"},
			{"practitioners", bsonPractitionerLinksMap},
			{"attachments", bsonAttachmentArrays},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"data", bsonInsolvency},
		}))

		mongoService.db = mt.DB
		attachmentResourceDao, err := mongoService.GetAttachmentResources("transactionID")

		assert.Nil(t, err)
		assert.NotNil(t, attachmentResourceDao)
		assert.Equal(t, attachmentResourceDao[0].ID, "ID")
		assert.Equal(t, attachmentResourceDao[0].Type, "type")
		assert.Equal(t, attachmentResourceDao[0].Status, "status")

	})

	mt.Run("GetAttachmentResources runs with attachments Nil", func(mt *mtest.T) {
		bsonInsolvency := bson.D{
			{"company_number", "CompanyNumber"},
			{"case_type", "CaseType"},
			{"company_name", "CompanyName"},
			{"practitioners", bsonPractitionerLinksMap},
			{"attachments", nil},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},

			{"data", bsonInsolvency},
		}))

		mongoService.db = mt.DB
		attachmentResourceDao, err := mongoService.GetAttachmentResources("transactionID")

		assert.Nil(t, err)
		assert.NotNil(t, attachmentResourceDao)
	})

	mt.Run("GetAttachmentResources runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		_, err := mongoService.GetAttachmentResources("transactionID")

		assert.Equal(t, err.Error(), "(Name) Message")
	})
}

func TestUnitGetAttachmentFromInsolvencyResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("GetAttachmentFromInsolvencyResource runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		_, err := mongoService.GetAttachmentFromInsolvencyResource("transactionID", "fileID")

		assert.Equal(t, err.Error(), "(Name) Message")
	})

	mt.Run("GetAttachmentFromInsolvencyResource runs successfully", func(mt *mtest.T) {

		bsonPractitionerLinksMap := bson.M{
			"PractionerID1": "PractitionerLink1",
			"PractionerID2": "PractitionerLink2",
		}

		bsonDataAttachment := bson.M{
			"id":     "ID",
			"type":   "type",
			"status": "status",
		}

		bsonAttachmentArrays := bson.A{bsonDataAttachment}

		bsonInsolvency := bson.D{
			{"company_number", "CompanyNumber"},
			{"case_type", "CaseType"},
			{"company_name", "CompanyName"},
			{"practitioners", bsonPractitionerLinksMap},
			{"attachments", bsonAttachmentArrays},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mongoService.db = mt.DB
		attachmentResourceDao, err := mongoService.GetAttachmentFromInsolvencyResource("transactionID", "fileID")

		assert.Nil(t, err)
		assert.NotNil(t, attachmentResourceDao)
		assert.Equal(t, attachmentResourceDao.ID, "ID")
		assert.Equal(t, attachmentResourceDao.Type, "type")
		assert.Equal(t, attachmentResourceDao.Status, "status")

	})
}

func TestUnitDeleteAttachmentResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	bsonPractitionerLinksMap := bson.M{
		"PractionerID1": "PractitionerLink1",
		"PractionerID2": "PractitionerLink2",
	}

	bsonInsolvency := bson.D{
		{"company_number", "CompanyNumber"},
		{"case_type", "CaseType"},
		{"company_name", "CompanyName"},
		{"practitioners", bsonPractitionerLinksMap},
	}

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("DeleteAttachmentResource runs successfully", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
			bson.E{Key: "upserted", Value: 1},
		))

		mongoService.db = mt.DB
		code, err := mongoService.DeleteAttachmentResource("transactionID", "attachmentID")

		assert.Nil(t, err)
		assert.Equal(t, code, 204)

	})

	mt.Run("DeleteAttachmentResource runs with findone error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB

		_, err := mongoService.DeleteAttachmentResource("transactionID", "attachmentID")

		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
	})

	mt.Run("DeleteAttachmentResource runs with error on UpdateOne", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		code, err := mongoService.DeleteAttachmentResource("transactionID", "attachmentID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID] - could not delete attachment with id [attachmentID]")
		assert.Equal(t, code, 500)

	})

	mt.Run("DeleteAttachmentResource runs with zero ModifiedCount", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"value", bson.D{
				{"_id", expectedInsolvency.ID},
				{"transaction_id", expectedInsolvency.TransactionID},
				{"etag", expectedInsolvency.Data.Etag},
				{"kind", expectedInsolvency.Data.Kind},
				{"data", bsonInsolvency},
			}},
		})

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 0},
			bson.E{Key: "upserted", Value: 1},
		))

		mongoService.db = mt.DB
		code, err := mongoService.DeleteAttachmentResource("transactionID", "attachmentID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID] - attachment with id [attachmentID] not found")
		assert.Equal(t, code, 404)

	})

}

func TestUnitCreateResolutionResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	resolutionResourceDao := models.ResolutionResourceDao{}

	mt.Run("CreateResolutionResource runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		code, err := mongoService.CreateResolutionResource(&resolutionResourceDao, "transactionID")

		assert.Equal(t, code, 500)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
	})

	mt.Run("CreateResolutionResource runs successfully with findone", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", expectedInsolvency.Data},
		}))

		mongoService.db = mt.DB
		code, err := mongoService.CreateResolutionResource(&resolutionResourceDao, "transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
		assert.Equal(t, code, 500)
	})

	mt.Run("CreateResolutionResource runs with error on UpdateOne", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", expectedInsolvency.Data},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		code, err := mongoService.CreateResolutionResource(&resolutionResourceDao, "transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
		assert.Equal(t, code, 500)
	})

	mt.Run("CreateResolutionResource runs with successfully on UpdateOne", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"data", expectedInsolvency.Data},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"nModified", 1},
		})

		mongoService.db = mt.DB
		code, err := mongoService.CreateResolutionResource(&resolutionResourceDao, "transactionID")

		assert.Nil(t, err)
		assert.Equal(t, code, 201)
	})
}

func TestUnitCreateStatementOfAffairsResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	statementOfAffairsResourceDao := models.StatementOfAffairsResourceDao{}

	mt.Run("CreateStatementOfAffairsResource runs successfully with findone", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", expectedInsolvency.Data},
		}))

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"value", bson.D{
				{"_id", expectedInsolvency.ID},
				{"transaction_id", expectedInsolvency.TransactionID},
				{"etag", expectedInsolvency.Data.Etag},
				{"kind", expectedInsolvency.Data.Kind},
				{"data", expectedInsolvency.Data},
			}},
		})

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 0},
			bson.E{Key: "upserted", Value: 1},
		))

		mongoService.db = mt.DB
		code, err := mongoService.CreateStatementOfAffairsResource(&statementOfAffairsResourceDao, "transactionID")

		assert.Nil(t, err)
		assert.Equal(t, code, 201)
	})
	mt.Run("CreateStatementOfAffairsResource runs with error in FindOne", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		code, err := mongoService.CreateStatementOfAffairsResource(&statementOfAffairsResourceDao, "transactionID")

		assert.Equal(t, code, 500)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
	})

	mt.Run("CreateStatementOfAffairsResource runs with error on UpdateOne", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", expectedInsolvency.Data},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		code, err := mongoService.CreateStatementOfAffairsResource(&statementOfAffairsResourceDao, "transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
		assert.Equal(t, code, 500)
	})

	mt.Run("CreateStatementOfAffairsResource with successfully on UpdateOne", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", expectedInsolvency.Data},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"nModified", 1},
		})

		mongoService.db = mt.DB
		code, err := mongoService.CreateStatementOfAffairsResource(&statementOfAffairsResourceDao, "transactionID")

		assert.Nil(t, err)
		assert.Equal(t, code, 201)
	})

}

func TestUnitGetStatementOfAffairsResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	bsonData := bson.M{
		"id":               "ID",
		"ip_code":          "IPCode",
		"first_name":       "FirstName",
		"last_name":        "LastName",
		"telephone_number": "TelephoneNumber",
		"email":            "Email",
	}

	bsonDataAttachment := bson.M{
		"id":     "ID",
		"type":   "type",
		"status": "status",
	}

	bsonArrays := bson.A{}
	bsonArrays = append(bsonArrays, bsonData)

	bsonAttachmentArrays := bson.A{}
	bsonAttachmentArrays = append(bsonAttachmentArrays, bsonDataAttachment)

	bsonStatementOfAffairsResourceDao := bson.D{
		{"statement_date", "statement_date"},
		{"attachments", []string{"attachments"}},
	}

	bsonPractitionerLinksMap := bson.M{
		"PractionerID1": "PractitionerLink1",
		"PractionerID2": "PractitionerLink2",
	}

	bsonInsolvency := bson.D{
		{"company_number", "CompanyNumber"},
		{"case_type", "CaseType"},
		{"company_name", "CompanyName"},
		{"practitioners", bsonPractitionerLinksMap},
		{"attachments", bsonAttachmentArrays},
		{"statement-of-affairs", bsonStatementOfAffairsResourceDao},
	}

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("GetStatementOfAffairsResource runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		_, err := mongoService.GetStatementOfAffairsResource("transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "(Name) Message")
	})

	mt.Run("GetStatementOfAffairsResource runs successfully with findone", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mongoService.db = mt.DB
		statementOfAffairsResourceDao, err := mongoService.GetStatementOfAffairsResource("transactionID")

		assert.Nil(t, err)
		assert.Equal(t, statementOfAffairsResourceDao.StatementDate, string("statement_date"))
		assert.Equal(t, statementOfAffairsResourceDao.Attachments[0], string("attachments"))
	})

	mt.Run("GetStatementOfAffairsResource - no insolvency case found", func(mt *mtest.T) {

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "models.InsolvencyResourceDao", mtest.FirstBatch))
		mongoService.db = mt.DB
		statementOfAffairsDao, err := mongoService.GetStatementOfAffairsResource("transactionID")

		assert.Equal(t, models.StatementOfAffairsResourceDao{}, statementOfAffairsDao)
		assert.Nil(t, err)
	})

	mt.Run("GetStatementOfAffairsResource - returned result can't be decoded", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{Key: "transaction_id", Value: bsonArrays},
		}))

		mongoService.db = mt.DB
		statementOfAffairsDao, err := mongoService.GetStatementOfAffairsResource("transactionID")

		assert.Equal(t, "error decoding key transaction_id: cannot decode array into a string type", err.Error())
		assert.Equal(t, models.StatementOfAffairsResourceDao{}, statementOfAffairsDao)
	})

	mt.Run("GetStatementOfAffairsResource - insolvency case contains no statement of affairs", func(mt *mtest.T) {

		bsonInsolvencyNoSOA := bson.D{
			{Key: "company_number", Value: "CompanyNumber"},
			{Key: "case_type", Value: "CaseType"},
			{Key: "company_name", Value: "CompanyName"},
			{Key: "practitioners", Value: bsonPractitionerLinksMap},
			{Key: "attachments", Value: bsonAttachmentArrays},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: expectedInsolvency.ID},
			{Key: "transaction_id", Value: expectedInsolvency.TransactionID},
			{Key: "etag", Value: expectedInsolvency.Data.Etag},
			{Key: "kind", Value: expectedInsolvency.Data.Kind},
			{Key: "data", Value: bsonInsolvencyNoSOA},
		}))

		mongoService.db = mt.DB
		statementOfAffairsDao, err := mongoService.GetStatementOfAffairsResource("transactionID")

		assert.Nil(t, err)
		assert.Equal(t, models.StatementOfAffairsResourceDao{}, statementOfAffairsDao)
	})
}

func TestUnitDeleteStatementOfAffairsResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	bsonPractitionerLinksMap := bson.M{
		"PractionerID1": "PractitionerLink1",
		"PractionerID2": "PractitionerLink2",
	}

	bsonInsolvency := bson.D{
		{"company_number", "CompanyNumber"},
		{"case_type", "CaseType"},
		{"company_name", "CompanyName"},
		{"practitioners", bsonPractitionerLinksMap},
	}

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("DeleteStatementOfAffairsResource runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB

		_, err := mongoService.DeleteStatementOfAffairsResource("transactionID")

		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
	})

	mt.Run("DeleteStatementOfAffairsResource runs with error on UpdateOne", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		code, err := mongoService.DeleteStatementOfAffairsResource("transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID] - could not delete statement of affairs")
		assert.Equal(t, code, 500)

	})

	mt.Run("DeleteStatementOfAffairsResource runs with zero ModifiedCount", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 0},
			bson.E{Key: "upserted", Value: 1},
		))

		mongoService.db = mt.DB
		code, err := mongoService.DeleteStatementOfAffairsResource("transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID] - statement of affairs not found")
		assert.Equal(t, code, 404)

	})

	mt.Run("DeleteStatementOfAffairsResource runs successfully", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
			bson.E{Key: "upserted", Value: 1},
		))

		mongoService.db = mt.DB
		code, err := mongoService.DeleteStatementOfAffairsResource("transactionID")

		assert.Nil(t, err)
		assert.Equal(t, code, 204)

	})
}

func TestUnitCreateProgressReportResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	progressReportResourceDao := models.ProgressReportResourceDao{}

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("CreateProgressReportResource with error findone", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		_, err := mongoService.CreateProgressReportResource(&progressReportResourceDao, "transactionID")

		assert.NotNil(t, err.Error())
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
	})

	mt.Run("CreateProgressReportResource with successful created one", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
			bson.E{Key: "upserted", Value: 1},
		))

		mongoService.db = mt.DB
		code, err := mongoService.CreateProgressReportResource(&progressReportResourceDao, "transactionID")

		assert.Nil(t, err)
		assert.Equal(t, code, 201)
	})
}

func TestUnitGetProgressReportResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("GetProgressReportResource runs successfully", func(mt *mtest.T) {
		bsonprogressReport := bson.D{
			{"from_date", "from_date"},
			{"to_date", "to_date"},
			{"attachments", []string{"attachments"}},
		}
		bsonInsolvencyResourceDaoData := bson.D{
			{"company_number", "company_number"},
			{"case_type", "case_type"},
			{"company_name", "company_name"},
			{"progress-report", bsonprogressReport},
		}

		bsonLink := bson.D{
			{"self", "self"},
		}

		bsonProgress := bson.D{
			{"transaction_id", "transaction_id"},
			{"attachments", "attachments"},
			{"etag", "etag"},
			{"kind", "kind"},
			{"data", bsonInsolvencyResourceDaoData},
			{"links", bsonLink},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.ProgressReportResourceDao", mtest.FirstBatch, bsonProgress))

		mongoService.db = mt.DB
		progressReportResource, err := mongoService.GetProgressReportResource("transactionID")

		assert.Nil(t, err)
		assert.Equal(t, progressReportResource.FromDate, string("from_date"))
		assert.Equal(t, progressReportResource.ToDate, string("to_date"))
		assert.Equal(t, progressReportResource.Attachments[0], string("attachments"))
	})

	mt.Run("GetProgressReportResource fails on wrong attachment", func(mt *mtest.T) {
		bsonprogressReport := bson.D{
			{"from_date", "from_date"},
			{"to_date", "to_date"},
			{"attachments", "attachments"},
		}
		bsonInsolvencyResourceDaoData := bson.D{
			{"company_number", "company_number"},
			{"case_type", "case_type"},
			{"company_name", "company_name"},
			{"progress-report", bsonprogressReport},
		}

		bsonLink := bson.D{
			{"self", "self"},
		}

		bsonProgress := bson.D{
			{"transaction_id", "transaction_id"},
			{"attachments", "attachments"},
			{"etag", "etag"},
			{"kind", "kind"},
			{"data", bsonInsolvencyResourceDaoData},
			{"links", bsonLink},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.ProgressReportResourceDao", mtest.FirstBatch, bsonProgress))

		mongoService.db = mt.DB
		progressReportResource, err := mongoService.GetProgressReportResource("transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, &models.ProgressReportResourceDao{}, progressReportResource)
	})

	mt.Run("GetProgressReportResource runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		progressReportResource, err := mongoService.GetProgressReportResource("transactionID")

		assert.Equal(t, err.Error(), "(Name) Message")
		assert.Equal(t, &models.ProgressReportResourceDao{}, progressReportResource)
	})

	mt.Run("GetProgressReportResource - no insolvency case found", func(mt *mtest.T) {

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "models.InsolvencyResourceDao", mtest.FirstBatch))
		mongoService.db = mt.DB
		progressReportDao, err := mongoService.GetProgressReportResource("transactionID")

		assert.Equal(t, &models.ProgressReportResourceDao{}, progressReportDao)
		assert.Nil(t, err)
	})

	mt.Run("GetProgressReportResource - insolvency case contains no progress report", func(mt *mtest.T) {

		bsonInsolvencyResourceDaoData := bson.D{
			{Key: "company_number", Value: "company_number"},
			{Key: "case_type", Value: "case_type"},
			{Key: "company_name", Value: "company_name"},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: expectedInsolvency.ID},
			{Key: "transaction_id", Value: expectedInsolvency.TransactionID},
			{Key: "etag", Value: expectedInsolvency.Data.Etag},
			{Key: "kind", Value: expectedInsolvency.Data.Kind},
			{Key: "links", Value: expectedInsolvency.Data.Links},
			{Key: "data", Value: bsonInsolvencyResourceDaoData},
		}))

		mongoService.db = mt.DB
		progressReportDao, err := mongoService.GetProgressReportResource("transactionID")

		assert.Nil(t, err)
		assert.Equal(t, &models.ProgressReportResourceDao{}, progressReportDao)
	})

}

func TestUnitDeleteProgressReportResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	bsonData := bson.M{
		"id":               "ID",
		"ip_code":          "IPCode",
		"first_name":       "FirstName",
		"last_name":        "LastName",
		"telephone_number": "TelephoneNumber",
		"email":            "Email",
	}

	bsonArrays := bson.A{}
	bsonArrays = append(bsonArrays, bsonData)
	bsonInsolvency := bson.D{
		{"company_number", "CompanyNumber"},
		{"case_type", "CaseType"},
		{"company_name", "CompanyName"},
		{"practitioners", bsonArrays},
	}

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("DeleteProgressReportResource runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB

		_, err := mongoService.DeleteProgressReportResource("transactionID")

		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
	})

	mt.Run("DeleteProgressReportResource runs with error on UpdateOne", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		code, err := mongoService.DeleteProgressReportResource("transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID] - could not delete progress report")
		assert.Equal(t, code, 500)

	})

	mt.Run("DeleteProgressReportResource runs with zero ModifiedCount", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 0},
			bson.E{Key: "upserted", Value: 1},
		))

		mongoService.db = mt.DB
		code, err := mongoService.DeleteProgressReportResource("transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID] - progress report not found")
		assert.Equal(t, code, 404)

	})

	mt.Run("DeleteProgressReportResource runs successfully", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
			bson.E{Key: "upserted", Value: 1},
		))

		mongoService.db = mt.DB
		code, err := mongoService.DeleteProgressReportResource("transactionID")

		assert.Nil(t, err)
		assert.Equal(t, code, 204)

	})
}

func TestUnitGetResolutionResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	bsonData := bson.M{
		"id":               "ID",
		"ip_code":          "IPCode",
		"first_name":       "FirstName",
		"last_name":        "LastName",
		"telephone_number": "TelephoneNumber",
		"email":            "Email",
	}

	bsonDataAttachment := bson.M{
		"id":     "ID",
		"type":   "type",
		"status": "status",
	}

	bsonArrays := bson.A{}
	bsonArrays = append(bsonArrays, bsonData)

	bsonAttachmentArrays := bson.A{}
	bsonAttachmentArrays = append(bsonAttachmentArrays, bsonDataAttachment)

	bsonStatementOfAffairsResourceDao := bson.D{
		{"statement_date", "statement_date"},
		{"attachments", []string{"attachments"}},
	}

	bsonResolution := bson.D{
		{"etag", "etag"},
		{"kind", "kind"},
		{"date_of_resolution", "date_of_resolution"},
		{"attachments", []string{"attachments"}},
	}

	bsonPractitionerLinksMap := bson.M{
		"PractionerID1": "PractitionerLink1",
		"PractionerID2": "PractitionerLink2",
	}

	bsonInsolvency := bson.D{
		{"company_number", "CompanyNumber"},
		{"case_type", "CaseType"},
		{"company_name", "CompanyName"},
		{"practitioners", bsonPractitionerLinksMap},
		{"attachments", bsonAttachmentArrays},
		{"resolution", bsonResolution},
		{"statement-of-affairs", bsonStatementOfAffairsResourceDao},
	}

	bsonInsolvencyNoResolution := bson.D{
		{Key: "company_number", Value: "CompanyNumber"},
		{Key: "case_type", Value: "CaseType"},
		{Key: "company_name", Value: "CompanyName"},
		{Key: "practitioners", Value: bsonPractitionerLinksMap},
		{Key: "attachments", Value: bsonAttachmentArrays},
		{Key: "statement-of-affairs", Value: bsonStatementOfAffairsResourceDao},
	}

	mt.Run("GetResolutionResource runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		code, err := mongoService.GetResolutionResource("transactionID")

		assert.NotNil(t, code)
		assert.Equal(t, err.Error(), "(Name) Message")
	})

	mt.Run("GetResolutionResource - no insolvency case found", func(mt *mtest.T) {

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "models.InsolvencyResourceDao", mtest.FirstBatch))
		mongoService.db = mt.DB
		resolutionDao, err := mongoService.GetResolutionResource("transactionID")

		assert.Equal(t, models.ResolutionResourceDao{}, resolutionDao)
		assert.Nil(t, err)
	})

	mt.Run("GetResolutionResource - returned result can't be decoded", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{Key: "transaction_id", Value: bsonArrays},
		}))

		mongoService.db = mt.DB
		resolutionDao, err := mongoService.GetResolutionResource("transactionID")

		assert.Equal(t, "error decoding key transaction_id: cannot decode array into a string type", err.Error())
		assert.Equal(t, models.ResolutionResourceDao{}, resolutionDao)
	})

	mt.Run("GetResolutionResource - insolvency case contains no resolution", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: expectedInsolvency.ID},
			{Key: "transaction_id", Value: expectedInsolvency.TransactionID},
			{Key: "etag", Value: expectedInsolvency.Data.Etag},
			{Key: "kind", Value: expectedInsolvency.Data.Kind},
			{Key: "data", Value: bsonInsolvencyNoResolution},
		}))

		mongoService.db = mt.DB
		resolutionDao, err := mongoService.GetResolutionResource("transactionID")

		assert.Nil(t, err)
		assert.Equal(t, models.ResolutionResourceDao{}, resolutionDao)
	})

	mt.Run("GetResolutionResource runs successfully with findone", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mongoService.db = mt.DB
		resolutionDao, err := mongoService.GetResolutionResource("transactionID")

		assert.Nil(t, err)
		assert.NotNil(t, resolutionDao.DateOfResolution)
	})
}

func TestUnitDeleteResolutionResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	bsonData := bson.M{
		"id":               "ID",
		"ip_code":          "IPCode",
		"first_name":       "FirstName",
		"last_name":        "LastName",
		"telephone_number": "TelephoneNumber",
		"email":            "Email",
	}

	bsonArrays := bson.A{}
	bsonArrays = append(bsonArrays, bsonData)
	bsonInsolvency := bson.D{
		{"company_number", "CompanyNumber"},
		{"case_type", "CaseType"},
		{"company_name", "CompanyName"},
		{"practitioners", bsonArrays},
	}

	mt.Run("DeleteResolutionResource runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB

		_, err := mongoService.DeleteResolutionResource("transactionID")

		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
	})

	mt.Run("DeleteResolutionResource runs with error on UpdateOne", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		code, err := mongoService.DeleteResolutionResource("transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID] - could not delete resolution")
		assert.Equal(t, code, 500)

	})

	mt.Run("DeleteResolutionResource runs with zero ModifiedCount", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 0},
			bson.E{Key: "upserted", Value: 1},
		))

		mongoService.db = mt.DB
		code, err := mongoService.DeleteResolutionResource("transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID] - resolution not found")
		assert.Equal(t, code, 404)

	})

	mt.Run("DeleteResolutionResource runs successfully", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Data.Etag},
			{"kind", expectedInsolvency.Data.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
			bson.E{Key: "upserted", Value: 1},
		))

		mongoService.db = mt.DB
		code, err := mongoService.DeleteResolutionResource("transactionID")

		assert.Nil(t, err)
		assert.Equal(t, code, 204)

	})
}
