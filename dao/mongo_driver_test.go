package dao

import (
	"testing"

	"github.com/companieshouse/insolvency-api/config"
	"github.com/companieshouse/insolvency-api/models"

	gomock "github.com/golang/mock/gomock"

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

	mongoService := MongoService{
		db:             dataBase,
		CollectionName: cfg.MongoCollection,
	}

	commanError := mtest.CommandError{
		Code:    1,
		Message: "Message",
		Name:    "Name",
		Labels:  []string{"label1"},
	}

	practitionerResourceDao := models.PractitionerResourceDao{
		ID:              "ID",
		IPCode:          "IPCode",
		FirstName:       "FirstName",
		LastName:        "LastName",
		TelephoneNumber: "TelephoneNumber",
		Email:           "Email",
		Address:         models.AddressResourceDao{},
		Role:            "Role",
		Links:           models.PractitionerResourceLinksDao{},
		Appointment:     &models.AppointmentResourceDao{},
	}

	practitioners := []models.PractitionerResourceDao{}

	dataInsolvency := models.InsolvencyResourceDaoData{
		CompanyNumber:      "CompanyNumber",
		CaseType:           "CaseType",
		CompanyName:        "CompanyName",
		Practitioners:      append(practitioners, practitionerResourceDao),
		Attachments:        []models.AttachmentResourceDao{},
		Resolution:         &models.ResolutionResourceDao{},
		StatementOfAffairs: &models.StatementOfAffairsResourceDao{},
	}

	expectedInsolvency := models.InsolvencyResourceDao{
		ID:            primitive.NewObjectID(),
		TransactionID: "TransactionID",
		Etag:          "Etag",
		Kind:          "Kind",
		Data:          dataInsolvency,
		Links:         models.InsolvencyResourceLinksDao{},
	}
	opts := mtest.NewOptions().DatabaseName("databaseName").ClientType(mtest.Mock)

	return mongoService, commanError, expectedInsolvency, opts, append(practitioners, practitionerResourceDao)
}

func TestUnitCreateInsolvencyResourceDriver(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("CreateInsolvencyResource with error findone", func(mt *mtest.T) {
		mt.Parallel()
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		err, _ := mongoService.CreateInsolvencyResource(&expectedInsolvency)

		assert.NotNil(t, err.Error())
		assert.Equal(t, err.Error(), "there was a problem creating an insolvency case for this transaction id: (Name) Message")
	})

	mt.Run("CreateInsolvencyResource with successful created one", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
		}))

		mongoService.db = mt.DB
		err, _ := mongoService.CreateInsolvencyResource(&expectedInsolvency)

		assert.Equal(t, err.Error(), "an insolvency case already exists for this transaction id")
	})
}

func TestUnitGetInsolvencyResourceDriver(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("GetInsolvencyResource runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		_, err := mongoService.GetInsolvencyResource("transactionID")

		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction [transactionID]")
	})

	mt.Run("GetInsolvencyResource runs successfully", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
		}))

		mongoService.db = mt.DB
		insolvencyResource, err := mongoService.GetInsolvencyResource("transactionID")

		assert.Nil(t, err)
		assert.NotNil(t, insolvencyResource)
		assert.Equal(t, insolvencyResource.Etag, expectedInsolvency.Etag)
		assert.Equal(t, insolvencyResource.Kind, expectedInsolvency.Kind)
	})
}

func TestUnitCreatePractitionersResourceDriver(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	practitionerResourceDao := models.PractitionerResourceDao{}

	mt.Run("CreatePractitionersResource runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		err, code := mongoService.CreatePractitionersResource(&practitionerResourceDao, "transactionID")

		assert.Equal(t, code, 500)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction transactionID")
	})

	mt.Run("CreatePractitionersResource runs successfully with findone", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", expectedInsolvency.Data.Practitioners},
		}))

		mongoService.db = mt.DB
		err, code := mongoService.CreatePractitionersResource(&practitionerResourceDao, "transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction transactionID")
		assert.Equal(t, code, 500)
	})
}

func TestUnitGetPractitionerResourceDriver(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, practitioners := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("GetPractitionerResource runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		practitioner, err := mongoService.GetPractitionerResource("practitionerID", "transactionID")

		assert.NotNil(t, practitioner)
		assert.Equal(t, err.Error(), "(Name) Message")
	})

	mt.Run("GetPractitionerResource runs successfully", func(mt *mtest.T) {
		mt.Parallel()

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

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mongoService.db = mt.DB
		insolvencyResource, err := mongoService.GetPractitionerResource("practitionerID", "transactionID")

		assert.Nil(t, err)
		assert.NotNil(t, insolvencyResource)
		assert.Equal(t, insolvencyResource.ID, practitioners[0].ID)
		assert.Equal(t, insolvencyResource.IPCode, practitioners[0].IPCode)
		assert.Equal(t, insolvencyResource.FirstName, practitioners[0].FirstName)
		assert.Equal(t, insolvencyResource.LastName, practitioners[0].LastName)
		assert.Equal(t, insolvencyResource.TelephoneNumber, practitioners[0].TelephoneNumber)
		assert.Equal(t, insolvencyResource.Email, practitioners[0].Email)
	})
}

func TestUnitGetPractitionerResourcesDriver(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, practitioners := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("GetPractitionerResources runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		_, err := mongoService.GetPractitionerResources("transactionID")

		assert.Equal(t, err.Error(), "(Name) Message")
	})

	mt.Run("GetPractitionerResource runs successfully", func(mt *mtest.T) {
		mt.Parallel()

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

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mongoService.db = mt.DB
		insolvencyResource, err := mongoService.GetPractitionerResources("transactionID")

		assert.Nil(t, err)
		assert.NotNil(t, insolvencyResource)
		assert.Equal(t, insolvencyResource[0].ID, practitioners[0].ID)
		assert.Equal(t, insolvencyResource[0].IPCode, practitioners[0].IPCode)
		assert.Equal(t, insolvencyResource[0].FirstName, practitioners[0].FirstName)
		assert.Equal(t, insolvencyResource[0].LastName, practitioners[0].LastName)
		assert.Equal(t, insolvencyResource[0].TelephoneNumber, practitioners[0].TelephoneNumber)
		assert.Equal(t, insolvencyResource[0].Email, practitioners[0].Email)

	})
}

func TestUnitDeletePractitionerDriver(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("DeletePractitioner runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		err, _ := mongoService.DeletePractitioner("practitionerID", "transactionID")

		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id transactionID")
	})

	mt.Run("DeletePractitioner runs successfully", func(mt *mtest.T) {
		mt.Parallel()

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

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"nModified", 1},
		})

		mongoService.db = mt.DB
		err, code := mongoService.DeletePractitioner("practitionerID", "transactionID")

		assert.Nil(t, err)
		assert.Equal(t, code, 204)

	})
}

func TestUnitAppointPractitionerDriver(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()
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

	appointmentResource := models.AppointmentResourceDao{
		AppointedOn: "AppointedOn",
		MadeBy:      "MadeBy",
		Links:       models.AppointmentResourceLinksDao{},
	}

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("AppointPractitioner runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB

		err, _ := mongoService.AppointPractitioner(&appointmentResource, "transactionID", "practitionerID")

		assert.Equal(t, err.Error(), "could not update practitioner appointment for practitionerID practitionerID: (Name) Message")
	})

	mt.Run("AppointPractitioner runs with zero MatchedCount", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"nModified", 1},
		})

		mongoService.db = mt.DB
		err, code := mongoService.AppointPractitioner(&appointmentResource, "practitionerID", "transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "item with transaction id practitionerID or practitioner id transactionID does not exist")
		assert.Equal(t, code, 404)

	})

	mt.Run("AppointPractitioner runs with zero ModifiedCount", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 0},
			bson.E{Key: "upserted", Value: 1},
		))

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		err, code := mongoService.AppointPractitioner(&appointmentResource, "practitionerID", "transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "item with transaction id practitionerID or practitioner id transactionID not updated")
		assert.Equal(t, code, 404)

	})

	mt.Run("AppointPractitioner runs successfully", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
			bson.E{Key: "upserted", Value: 1},
		))

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		err, code := mongoService.AppointPractitioner(&appointmentResource, "practitionerID", "transactionID")

		assert.Nil(t, err)
		assert.Equal(t, code, 204)

	})
}

func TestUnitDeletePractitionerAppointmentDriver(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

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

	mt.Run("DeletePractitionerAppointment runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB

		err, _ := mongoService.DeletePractitionerAppointment("transactionID", "practitionerID")

		assert.Equal(t, err.Error(), "could not update practitioner appointment for practitionerID practitionerID: (Name) Message")
	})

	mt.Run("DeletePractitionerAppointment runs with zero MatchedCount", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"nModified", 1},
		})

		mongoService.db = mt.DB
		err, code := mongoService.DeletePractitionerAppointment("practitionerID", "transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "item with transaction id practitionerID or practitioner id transactionID does not exist")
		assert.Equal(t, code, 404)

	})

	mt.Run("DeletePractitionerAppointment runs with zero ModifiedCount", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 0},
			bson.E{Key: "upserted", Value: 1},
		))

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		err, code := mongoService.DeletePractitionerAppointment("practitionerID", "transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "item with transaction id practitionerID or practitioner id transactionID not updated")
		assert.Equal(t, code, 404)

	})

	mt.Run("DeletePractitionerAppointment runs successfully", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
			bson.E{Key: "upserted", Value: 1},
		))

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		err, code := mongoService.DeletePractitionerAppointment("practitionerID", "transactionID")

		assert.Nil(t, err)
		assert.Equal(t, code, 204)

	})
}

func TestUnitAddAttachmentToInsolvencyResourceDriver(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("AddAttachmentToInsolvencyResource runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		_, err := mongoService.AddAttachmentToInsolvencyResource("transactionID", "fileID", "attachmentType")

		assert.Equal(t, err.Error(), "error updating mongo for transaction [transactionID]: [(Name) Message]")
	})

	mt.Run("AddAttachmentToInsolvencyResource runs successfully with findone", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
			bson.E{Key: "upserted", Value: 1},
		))

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", expectedInsolvency.Data.Practitioners},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		attachmentDao, err := mongoService.AddAttachmentToInsolvencyResource("transactionID", "fileID", "attachmentType")

		assert.Nil(t, err)
		assert.NotNil(t, attachmentDao)
		assert.Equal(t, attachmentDao.ID, "fileID")
		assert.Equal(t, attachmentDao.Type, "attachmentType")
		assert.Equal(t, attachmentDao.Status, "submitted")
	})
}

func TestUnitGetAttachmentResourcesDriver(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("GetAttachmentResources runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		_, err := mongoService.GetAttachmentResources("transactionID")

		assert.Equal(t, err.Error(), "(Name) Message")
	})

	mt.Run("GetAttachmentResources runs successfully", func(mt *mtest.T) {
		mt.Parallel()

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
			{"practitioners", bsonArrays},
			{"attachments", bsonAttachmentArrays},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
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
}

func TestUnitGetAttachmentFromInsolvencyResourceDriver(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("GetAttachmentFromInsolvencyResource runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		_, err := mongoService.GetAttachmentFromInsolvencyResource("transactionID", "fileID")

		assert.Equal(t, err.Error(), "(Name) Message")
	})

	mt.Run("GetAttachmentFromInsolvencyResource runs successfully", func(mt *mtest.T) {
		mt.Parallel()

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
			{"practitioners", bsonArrays},
			{"attachments", bsonAttachmentArrays},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
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
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

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

	mt.Run("DeleteAttachmentResource runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB

		_, err := mongoService.DeleteAttachmentResource("transactionID", "attachmentID")

		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
	})

	mt.Run("DeleteAttachmentResource runs with zero ModifiedCount", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

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

	mt.Run("DeleteAttachmentResource runs successfully", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

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
}

func TestUnitUpdateAttachmentStatusDriver(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

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
		{"practitioners", bsonArrays},
		{"attachments", bsonAttachmentArrays},
	}

	mt.Run("UpdateAttachmentStatus runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB

		_, err := mongoService.UpdateAttachmentStatus("transactionID", "attachmentID", "avStatus")

		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
	})

	mt.Run("UpdateAttachmentStatus runs with error on UpdateOne", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mongoService.db = mt.DB
		code, err := mongoService.UpdateAttachmentStatus("transactionID", "attachmentID", "avStatus")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID] - could not update status of attachment with id [attachmentID]")
		assert.Equal(t, code, 500)

	})

	mt.Run("UpdateAttachmentStatus runs successfully with status not processed", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
			bson.E{Key: "upserted", Value: 1},
		))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		code, err := mongoService.UpdateAttachmentStatus("transactionID", "attachmentID", "avStatus")

		assert.NotNil(t, err)
		assert.Equal(t, code, 500)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID] - could not update status of attachment with id [attachmentID]")
	})

	mt.Run("UpdateAttachmentStatus runs successfully with ModifiedCount zero", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

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

	mt.Run("UpdateAttachmentStatus runs successfully", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

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
}

func TestUnitCreateResolutionResourceDriver(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	resolutionResourceDao := models.ResolutionResourceDao{}

	mt.Run("CreateResolutionResource runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		code, err := mongoService.CreateResolutionResource(&resolutionResourceDao, "transactionID")

		assert.Equal(t, code, 500)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction transactionID")
	})

	mt.Run("CreateResolutionResource runs successfully with findone", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", expectedInsolvency.Data.Practitioners},
		}))

		mongoService.db = mt.DB
		code, err := mongoService.CreateResolutionResource(&resolutionResourceDao, "transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction transactionID")
		assert.Equal(t, code, 500)
	})
}

func TestUnitCreateStatementOfAffairsResourceDriver(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

	mt := mtest.New(t, opts)
	defer mt.Close()

	statementOfAffairsResourceDao := models.StatementOfAffairsResourceDao{}

	mt.Run("CreateStatementOfAffairsResource runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		code, err := mongoService.CreateStatementOfAffairsResource(&statementOfAffairsResourceDao, "transactionID")

		assert.Equal(t, code, 500)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction transactionID")
	})

	mt.Run("CreateStatementOfAffairsResource runs successfully with findone", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", expectedInsolvency.Data.Practitioners},
		}))

		mongoService.db = mt.DB
		code, err := mongoService.CreateStatementOfAffairsResource(&statementOfAffairsResourceDao, "transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction transactionID")
		assert.Equal(t, code, 500)
	})
}

func TestUnitGetStatementOfAffairsResourceDriver(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

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
		{"practitioners", bsonArrays},
		{"attachments", bsonAttachmentArrays},
		{"statement-of-affairs", bsonStatementOfAffairsResourceDao},
	}

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("GetStatementOfAffairsResource runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		_, err := mongoService.GetStatementOfAffairsResource("transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "(Name) Message")
	})

	mt.Run("GetStatementOfAffairsResource runs successfully with findone", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
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
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

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
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB

		_, err := mongoService.DeleteStatementOfAffairsResource("transactionID")

		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
	})

	mt.Run("DeleteStatementOfAffairsResource runs with error on UpdateOne", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		code, err := mongoService.DeleteStatementOfAffairsResource("transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID] - could not delete statement of affairs")
		assert.Equal(t, code, 500)

	})

	mt.Run("DeleteStatementOfAffairsResource runs with zero ModifiedCount", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

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
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

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

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

	progressReportResourceDao := models.ProgressReportResourceDao{}

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("CreateProgressReportResource with error findone", func(mt *mtest.T) {
		mt.Parallel()
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		_, err := mongoService.CreateProgressReportResource(&progressReportResourceDao,"transactionID")

		assert.NotNil(t, err.Error())
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction transactionID")
	})

	mt.Run("CreateProgressReportResource with successful created one", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

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

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

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

	bsonResolution:=bson.D{
		{"etag", "etag"},
		{"kind", "kind"},
		{"date_of_resolution", "date_of_resolution"},
		{"attachments", []string{"attachments"}},
	}

	bsonInsolvency := bson.D{
		{"company_number", "CompanyNumber"},
		{"case_type", "CaseType"},
		{"company_name", "CompanyName"},
		{"practitioners", bsonArrays},
		{"attachments", bsonAttachmentArrays},
		{"resolution", bsonResolution},
		{"statement-of-affairs", bsonStatementOfAffairsResourceDao},
	}

	mt := mtest.New(t, opts)
	defer mt.Close()

	mt.Run("GetResolutionResource runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		code, err := mongoService.GetResolutionResource("transactionID")

		assert.NotNil(t, code)
		assert.Equal(t, err.Error(), "(Name) Message")
	})

	mt.Run("GetResolutionResource runs successfully with findone", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

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
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mongoService, commanError, expectedInsolvency, opts, _ := setDriverUp()

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

	mt.Run("DeleteResolutionResource runs with error", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB

		_, err := mongoService.DeleteResolutionResource("transactionID")

		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID]")
	})

	mt.Run("DeleteResolutionResource runs with error on UpdateOne", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

		mongoService.db = mt.DB
		code, err := mongoService.DeleteResolutionResource("transactionID")

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "there was a problem handling your request for transaction id [transactionID] - could not delete resolution")
		assert.Equal(t, code, 500)

	})

	mt.Run("DeleteResolutionResource runs with zero ModifiedCount", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

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
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "models.InsolvencyResourceDao", mtest.FirstBatch, bson.D{
			{"_id", expectedInsolvency.ID},
			{"transaction_id", expectedInsolvency.TransactionID},
			{"etag", expectedInsolvency.Etag},
			{"kind", expectedInsolvency.Kind},
			{"data", bsonInsolvency},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(commanError))

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
