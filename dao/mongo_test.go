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

func TestUnitGetInsolvencyPractitionersResource(t *testing.T) {

	Convey("Get Insolvency Resource", t, func() {

		mongoService := setUp(t)

		_, _, err := mongoService.GetInsolvencyPractitionersResource("transactionID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction [transactionID]")
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

func TestUnitGetInsolvencyPractitionerByTransactionID(t *testing.T) {

	Convey("Get practitioner resources", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.GetInsolvencyResourceData("transactionID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction transactionID")
	})
}

func TestUnitGetPractitionersByIds(t *testing.T) {

	Convey("Get practitioner resources", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.GetPractitionersAppointmentResource([]string{"practitionerID"}, "transactionID")

		So(err.Error(), ShouldEqual, "the Aggregate operation must have a Deployment set before Execute can be called")
	})
}

func TestUnitDeletePractitioner(t *testing.T) {

	Convey("Delete practitioner", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.DeletePractitioner("practitionerID", "transactionID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction id transactionID")
	})
}

func TestUnitDeletePractitionerAppointment(t *testing.T) {

	Convey("Delete practitioner appointment", t, func() {

		mongoService := setUp(t)

		_, err := mongoService.DeletePractitionerAppointment("transactionID", "practitionerID")

		So(err.Error(), ShouldEqual, "could not update practitioner appointment for practitionerID practitionerID: the Update operation must have a Deployment set before Execute can be called")
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

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction transactionID")
	})
}

func TestUnitCreateStatementOfAffairsResource(t *testing.T) {

	Convey("Create statement of affairs resource", t, func() {

		mongoService := setUp(t)

		statementResource := models.StatementOfAffairsResourceDao{}

		_, err := mongoService.CreateStatementOfAffairsResource(&statementResource, "transactionID")

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction transactionID")
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

		So(err.Error(), ShouldEqual, "there was a problem handling your request for transaction transactionID")
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
