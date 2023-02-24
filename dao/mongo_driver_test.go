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

	jsonPractitionersDao := `{
		"VM04221441": "/transactions/168570-809316-704268/insolvency/practitioners/VM04221441"
	}`

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
	dataInsolvency.Data.Practitioners = jsonPractitionersDao
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

	bsonInsolvency := bson.D{
		{"company_number", "CompanyNumber"},
		{"case_type", "CaseType"},
		{"company_name", "CompanyName"},
		{"practitioners", "bsonArrays"},
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

func TestUnitGetPractitionersAppointmentResourceDriver(t *testing.T) {
	t.Parallel()

	mongoService, commandError, _, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("GetPractitionersAppointmentResource runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		practitioner, err := mongoService.GetPractitionersAppointmentResource([]string{"practitionerID"}, "transactionID")

		assert.Nil(t, practitioner)
		assert.Equal(t, err.Error(), "(Name) Message")
	})

	mt.Run("GetPractitionersAppointmentResource runs successfully", func(mt *mtest.T) {
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
		insolvencyResource, err := mongoService.GetPractitionersAppointmentResource([]string{"practitionerID"}, "transactionID")

		assert.Nil(t, err)
		assert.NotNil(t, insolvencyResource)
	})
}

func TestUnitDeletePractitionerDriver(t *testing.T) {
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

	mt.Run("DeletePractitioner runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		_, err := mongoService.DeletePractitioner("practitionerID", "transactionID")

		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id transactionID")
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

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"nModified", 0},
		})

		mongoService.db = mt.DB
		code, err := mongoService.DeletePractitioner("practitionerID", "transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id transactionID - practitioner with id practitionerID not found")
		assert.Equal(t, code, 404)

	})

	mt.Run("DeletePractitioner runs successfully", func(mt *mtest.T) {
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
		code, err := mongoService.DeletePractitioner("practitionerID", "transactionID")

		assert.Nil(t, err)
		assert.Equal(t, code, 204)

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
		_, err := mongoService.UpdatePractitionerAppointment(&appointmentResourceDao, "transactionID", "practitionerID")

		assert.Equal(t, err.Error(), "could not update practitioner appointment for practitionerID practitionerID: (Name) Message")
	})
}

func TestUnitDeletePractitionerAppointmentDriver(t *testing.T) {
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

	mt.Run("DeletePractitionerAppointment runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB

		_, err := mongoService.DeletePractitionerAppointment("transactionID", "practitionerID")

		assert.Equal(t, err.Error(), "could not update practitioner appointment for practitionerID practitionerID: (Name) Message")
	})

	mt.Run("DeletePractitionerAppointment runs successfully", func(mt *mtest.T) {
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
		code, err := mongoService.DeletePractitionerAppointment("practitionerID", "transactionID")

		assert.Nil(t, err)
		assert.Equal(t, code, 204)

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

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("GetAttachmentResources runs successfully", func(mt *mtest.T) {
		bsonInsolvency := bson.D{
			{"company_number", "CompanyNumber"},
			{"case_type", "CaseType"},
			{"company_name", "CompanyName"},
			{"practitioners", "bsonArrays"},
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
			{"practitioners", "bsonArrays"},
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

		bsonInsolvency := bson.D{
			{"company_number", "CompanyNumber"},
			{"case_type", "CaseType"},
			{"company_name", "CompanyName"},
			{"practitioners", "bsonArrays"},
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
		{"practitioners", "bsonArrays"},
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
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction transactionID")
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
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction transactionID")
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
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction transactionID")
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
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction transactionID")
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
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction transactionID")
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

	bsonInsolvency := bson.D{
		{"company_number", "CompanyNumber"},
		{"case_type", "CaseType"},
		{"company_name", "CompanyName"},
		{"practitioners", "bsonArrays"},
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
}

func TestUnitDeleteStatementOfAffairsResourceDriver(t *testing.T) {
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
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction transactionID")
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

	bsonInsolvency := bson.D{
		{"company_number", "CompanyNumber"},
		{"case_type", "CaseType"},
		{"company_name", "CompanyName"},
		{"practitioners", "bsonArrays"},
		{"attachments", bsonAttachmentArrays},
		{"resolution", bsonResolution},
		{"statement-of-affairs", bsonStatementOfAffairsResourceDao},
	}

	mt.Run("GetResolutionResource runs with error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commandError))

		mongoService.db = mt.DB
		code, err := mongoService.GetResolutionResource("transactionID")

		assert.NotNil(t, code)
		assert.Equal(t, err.Error(), "(Name) Message")
	})

	mt.Run("GetResolutionResource runs successfully with findone", func(mt *mtest.T) {
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
		resolutionDao, err := mongoService.GetResolutionResource("transactionID")

		assert.Nil(t, err)
		assert.NotNil(t, resolutionDao)
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
