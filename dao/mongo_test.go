package dao

import (
	"testing"

	"github.com/companieshouse/insolvency-api/config"
	"github.com/companieshouse/insolvency-api/models"
	gomock "github.com/golang/mock/gomock"

	"go.mongodb.org/mongo-driver/mongo"

	. "github.com/smartystreets/goconvey/convey"
)

func NewGetMongoDatabase(mongoDBURL, databaseName string) MongoDatabaseInterface {
	return getMongoClient(mongoDBURL).Database(databaseName)
}

func setUp(t *testing.T) MongoService {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	client = &mongo.Client{}
	cfg, _ := config.Get()
	dataBase := NewGetMongoDatabase("mongoDBURL", "databaseName")

	mongoService := MongoService{
		db:             dataBase,
		CollectionName: cfg.MongoCollection,
	}
	return mongoService
}

func TestUnitCreateInsolvencyResource(t *testing.T) {

	Convey("Create Insolvency Resource", t, func() {

		expectedInsolvency := models.InsolvencyResourceDao{}

		mongoService := setUp(t)

		_, err := mongoService.CreateInsolvencyResource(&expectedInsolvency)

		So(err.Error(), ShouldEqual, "there was a problem creating an insolvency case for this transaction id: the Find operation must have a Deployment set before Execute can be called")
	})
}

func TestUnitGetInsolvencyAndExpandedPractitionerResources(t *testing.T) {

	Convey("Get Insolvency Resource", t, func() {

		mongoService := setUp(t)

		_, _, err := mongoService.GetInsolvencyAndExpandedPractitionerResources("transactionID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction transactionID")
	})
}

func TestUnitGetInsolvencyResource(t *testing.T) {

	Convey("Get Insolvency Resource", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.GetInsolvencyResource("transactionID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction transactionID")
	})
}

func TestUnitCreatePractitionerResource(t *testing.T) {

	Convey("Create a practitioner resource", t, func() {

		mongoService := setUp(t)

		practitionerResource := models.PractitionerResourceDao{}

		_, err := mongoService.CreatePractitionerResource(&practitionerResource, "transactionID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction transactionID (insert practitioner to collection)")
	})
}

func TestUnitGetPractitionerAppointmentResource(t *testing.T) {

	Convey("Get practitioner resources", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.GetPractitionerAppointment("transactionID", "practitionerID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction id [transactionID]")
	})
}

func TestUnitGetSinglePractitionerResource(t *testing.T) {

	Convey("Get practitioner resources", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.GetSinglePractitionerResource("transactionID", "practitionerID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction id [transactionID]")
	})
}

func TestUnitDeletePractitioner(t *testing.T) {

	Convey("Delete practitioner", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.DeletePractitioner("transactionID", "practitionerID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction id transactionID")
	})
}

func TestUnitDeletePractitionerAppointment(t *testing.T) {

	Convey("Delete practitioner appointment", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.DeletePractitionerAppointment("transactionID", "practitionerID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction id [transactionID]")
	})
}

func TestUnitAddAttachmentToInsolvencyResource(t *testing.T) {

	Convey("Add attachment to insolvency resource", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.AddAttachmentToInsolvencyResource("transactionID", "fileID", "attachmentType")

		So(err.Error(), ShouldEqual, "error updating mongo for transaction [transactionID]: [the Update operation must have a Deployment set before Execute can be called]")
	})
}

func TestUnitGetAttachmentResources(t *testing.T) {

	Convey("Get attachment resources", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.GetAttachmentResources("transactionID")

		So(err.Error(), ShouldEqual, "the Find operation must have a Deployment set before Execute can be called")
	})
}

func TestUnitGetAttachmentFromInsolvencyResource(t *testing.T) {

	Convey("Get attachment from insolvency resource", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.GetAttachmentFromInsolvencyResource("transactionID", "fileID")

		So(err.Error(), ShouldEqual, "the Find operation must have a Deployment set before Execute can be called")
	})
}

func TestUnitDeleteAttachmentResource(t *testing.T) {

	Convey("Delete attachment status", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.DeleteAttachmentResource("transactionID", "attachmentID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction id [transactionID]")
	})
}

func TestUnitUpdateAttachmentStatus(t *testing.T) {

	Convey("Update attachment status", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.UpdateAttachmentStatus("transactionID", "attachmentID", "avStatus")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction id [transactionID]")
	})
}

func TestUnitCreateResolutionResource(t *testing.T) {

	Convey("Create resolution resource", t, func() {

		mongoService := setUp(t)

		resolutionResource := models.ResolutionResourceDao{}

		_, err := mongoService.CreateResolutionResource(&resolutionResource, "transactionID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction id [transactionID]")
	})
}

func TestUnitCreateStatementOfAffairsResource(t *testing.T) {

	Convey("Create statement of affairs resource", t, func() {

		mongoService := setUp(t)

		statementResource := models.StatementOfAffairsResourceDao{}

		_, err := mongoService.CreateStatementOfAffairsResource(&statementResource, "transactionID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction id [transactionID]")
	})
}

func TestUnitGetStatementOfAffairsResource(t *testing.T) {

	Convey("Get statement of affairs resource", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.GetStatementOfAffairsResource("transactionID")

		So(err.Error(), ShouldEqual, "the Find operation must have a Deployment set before Execute can be called")
	})
}

func TestUnitDeleteStatementOfAffairsResource(t *testing.T) {

	Convey("Delete statement of affairs resource", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.DeleteStatementOfAffairsResource("transactionID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction id [transactionID]")
	})
}

func TestUnitCreateProgressReportResource(t *testing.T) {

	Convey("Create progress report resource", t, func() {

		mongoService := setUp(t)

		progressReport := models.ProgressReportResourceDao{}

		_, err := mongoService.CreateProgressReportResource(&progressReport, "transactionID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction id [transactionID]")
	})
}

func TestUnitGetProgressReportResource(t *testing.T) {

	Convey("Get progress report resource", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.GetProgressReportResource("transactionID")

		So(err.Error(), ShouldEqual, "the Find operation must have a Deployment set before Execute can be called")
	})
}

func TestUnitDeleteProgressReportResource(t *testing.T) {

	Convey("Delete progress report", t, func() {

		MongoService := setUp(t)

		_, err := MongoService.DeleteProgressReportResource("transactionID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction id [transactionID]")

	})
}

func TestUnitGetResolutionResource(t *testing.T) {

	Convey("Get resolution resource", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.GetResolutionResource("transactionID")

		So(err.Error(), ShouldEqual, "the Find operation must have a Deployment set before Execute can be called")
	})
}

func TestUnitDeleteResolutionResource(t *testing.T) {
	Convey("Delete resolution resource", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.DeleteResolutionResource("transactionID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction id [transactionID]")
	})
}

func TestUnitDeleteResource(t *testing.T) {
	Convey("DeleteResource", t, func() {

		MongoService := setUp(t)

		_, err := MongoService.DeleteResource("transactionID", "progress-report")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction id [transactionID]")
	})
}
