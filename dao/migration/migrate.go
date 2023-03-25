package main

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/config"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// this can be run multiple times with the same output using the command go run ./dao/migration/migrate.go depending on your file structure.
// if error at any point it would stop without completing or distrupt or damage any file.
// please look at whatever error message it outputs and rectify, i hope it does not get to that based on the tests alrady done.
// A successfull migration would back up the insolvency to insolvency_backup collection,with the backup, if any issue(s) you can revert insolvency_bacup to original insolvency.
// once happy with your migration you may drop the insolvency_backup or keep for reference.
// please note some models are recreated in order to be able to complete the migration without distorting the documents.
// also note this script would only migrate practitioner and its appointments.However, could be extended on lines 363-365.

type InsolvencyMigrationDao struct {
	ID            primitive.ObjectID `bson:"_id"`
	TransactionID string             `bson:"transaction_id"`
	Etag          string             `bson:"etag"`
	Kind          string             `bson:"kind"`
	Data          struct {
		CompanyNumber      string                            `bson:"company_number"`
		CaseType           string                            `bson:"case_type"`
		CompanyName        string                            `bson:"company_name"`
		Practitioners      []PractitionerMigrationDao        `bson:"practitioners,omitempty"`
		Attachments        []AttachmentMigrationDao          `bson:"attachments,omitempty"`
		Resolution         *ResolutionMigrationDao           `bson:"resolution,omitempty"`
		StatementOfAffairs *StatementOfAffairsMigrationDao   `bson:"statement-of-affairs,omitempty"`
		ProgressReport     *models.ProgressReportResourceDao `bson:"progress-report,omitempty"`
	}
	Links models.InsolvencyResourceLinksDao `bson:"links,omitempty"`
}

type ResolutionMigrationDao struct {
	Etag             *string                      `bson:"etag,omitempty"`
	Kind             *string                      `bson:"kind,omitempty"`
	DateOfResolution *string                      `bson:"date_of_resolution,omitempty"`
	Attachments      *[]string                    `bson:"attachments,omitempty"`
	Links            *ResolutionMigrationLinksDao `bson:"links,omitempty"`
}

type ResolutionMigrationLinksDao struct {
	Self *string `bson:"self,omitempty"`
}

// StatementOfAffairsResourceDao contains the data for the statement of affairs DB resource
type StatementOfAffairsMigrationDao struct {
	Etag          string                                     `bson:"etag,omitempty"`
	Kind          string                                     `bson:"kind,omitempty"`
	StatementDate string                                     `bson:"statement_date,omitempty"`
	Attachments   []string                                   `bson:"attachments,omitempty"`
	Links         *models.StatementOfAffairsResourceLinksDao `bson:"links,omitempty"`
}

// PractitionerResourceDao contains the data for the practitioner resource in Mongo
type PractitionerMigrationDao struct {
	Id              string                       `bson:"id"`
	IPCode          string                       `bson:"ip_code"`
	FirstName       string                       `bson:"first_name"`
	LastName        string                       `bson:"last_name"`
	TelephoneNumber string                       `bson:"telephone_number,omitempty"`
	Email           string                       `bson:"email,omitempty"`
	Address         AddressMigrationDao          `bson:"address"`
	Role            string                       `bson:"role"`
	Links           PractitionerMigrationLinkDao `bson:"links"`
	Appointment     *AppointmentMigrationDao     `bson:"appointment,omitempty"`
}

// AddressResourceDao contains the data for any addresses in Mongo
type AddressMigrationDao struct {
	Premises     string `bson:"premises"`
	AddressLine1 string `bson:"address_line_1"`
	AddressLine2 string `bson:"address_line_2"`
	Country      string `bson:"country"`
	Locality     string `bson:"locality"`
	Region       string `bson:"region"`
	PostalCode   string `bson:"postal_code"`
	POBox        string `bson:"po_box"`
}

// AttachmentMigrationDao contains the data for the attachment DB resource
type AttachmentMigrationDao struct {
	ID     string                      `bson:"id"`
	Type   string                      `bson:"type"`
	Status string                      `bson:"status"`
	Links  AttachmentMigrationLinksDao `bson:"links"`
}

// AttachmentMigrationLinksDao contains the Links data for an attachment
type AttachmentMigrationLinksDao struct {
	Self     string `bson:"self"`
	Download string `bson:"download"`
}

// AppointmentMigrationDao contains the data for the appointment resource in Mongo
type AppointmentMigrationDao struct {
	AppointedOn string                       `bson:"appointed_on,omitempty"`
	MadeBy      string                       `bson:"made_by,omitempty"`
	Links       AppointmentMigrationLinksDao `bson:"links,omitempty"`
	Etag        string                       `bson:"etag,omitempty"`
	Kind        string                       `bson:"kind,omitempty"`
}

// AppointmentMigrationLinksDao contains the Links data for an appointment
type AppointmentMigrationLinksDao struct {
	Self string `bson:"self,omitempty"`
}

// PractitionerMigrationLinkDao contains the Links data for a practitioner
type PractitionerMigrationLinkDao struct {
	Self        string `bson:"self"`
	Appointment string `bson:"appointment,omitempty"`
}

var client *mongo.Client

func getMongoClient(mongoDBURL string) *mongo.Client {
	if client != nil {
		return client
	}

	ctx := context.Background()

	clientOptions := options.Client().ApplyURI(mongoDBURL)
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// check we can connect to the mongodb instance. failure here should result in a crash.
	pingContext, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
	defer cancel()
	err = client.Ping(pingContext, nil)
	if err != nil {
		log.Error(errors.New("ping to mongodb timed out. please check the connection to mongodb and that it is running"))
		os.Exit(1)
	}

	log.Info("connected to mongodb successfully")

	return client
}

type MongoMigrationService struct {
	db             MongoDatabaseInterface
	Client         *mongo.Client
	CollectionName string
}

// MongoDatabaseInterface is an interface that describes the mongodb driver
type MongoDatabaseInterface interface {
	Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection
}

func getMongoDatabase(mongoDBURL, databaseName string) MongoDatabaseInterface {
	return getMongoClient(mongoDBURL).Database(databaseName)
}

func NewDAOMigrationService(cfg *config.Config) MigrationService {
	database := getMongoDatabase(cfg.MongoDBURL, cfg.Database)
	return &MongoMigrationService{
		db:             database,
		CollectionName: cfg.MongoCollection,
	}
}

type MigrationService interface {
	Migrate() (*[]InsolvencyMigrationDao, error)
}

// generateEtag generates a random etag which is generated on every write action
func generateEtag() string {
	// Get a random number and the time in seconds and milliseconds
	timeNow := time.Now()
	rand.Seed(timeNow.UTC().UnixNano())

	// Calculate a SHA-512 truncated digest
	shaDigest := sha512.New512_224()
	sha1Hash := hex.EncodeToString(shaDigest.Sum(nil))

	return sha1Hash
}

func transform(insolvencyMigrationDaos []InsolvencyMigrationDao) []interface{} {
	var insolvencyFromBackup []interface{}
	for _, insolvency := range insolvencyMigrationDaos {
		insolvencyModel := InsolvencyMigrationDao{}
		insolvencyModel.ID = insolvency.ID
		insolvencyModel.Etag = insolvency.Etag
		insolvencyModel.Kind = insolvency.Kind
		insolvencyModel.TransactionID = insolvency.TransactionID
		insolvencyModel.Data.CompanyNumber = insolvency.Data.CompanyNumber
		insolvencyModel.Data.CaseType = insolvency.Data.CaseType
		insolvencyModel.Data.CompanyName = insolvency.Data.CompanyNumber
		insolvencyModel.Data.Attachments = insolvency.Data.Attachments
		insolvencyModel.Data.Practitioners = insolvency.Data.Practitioners
		insolvencyModel.Data.StatementOfAffairs = insolvency.Data.StatementOfAffairs
		insolvencyModel.Data.Resolution = insolvency.Data.Resolution
		insolvencyModel.Data.ProgressReport = insolvency.Data.ProgressReport
		insolvencyModel.Links = insolvency.Links

		insolvencyFromBackup = append(insolvencyFromBackup, insolvencyModel)
	}
	return insolvencyFromBackup
}

func (m *MongoMigrationService) Migrate() (*[]InsolvencyMigrationDao, error) {

	collection := m.db.Collection(m.CollectionName)
	practitionerCollectionName := m.db.Collection(dao.PractitionerCollectionName)
	appointmentCollectionName := m.db.Collection(dao.AppointmentCollectionName)
	insolvencyBackUpCollectionName := m.db.Collection("insolvency_backup")
	// drop previous created collections, change insolvency_backup to insolvency to start rerun

	err := practitionerCollectionName.Drop(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to drop practitioner collection")
	}

	//drop appoinment
	err = appointmentCollectionName.Drop(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to drop appointment collection")
	}

	// Retrieve insolvency case from insolvency_backup
	var insolvencyDaosFromBackup []InsolvencyMigrationDao
	storedInsolvencyBackup, _ := insolvencyBackUpCollectionName.Find(context.Background(), bson.M{})
	if storedInsolvencyBackup != nil {
		err = storedInsolvencyBackup.All(context.Background(), &insolvencyDaosFromBackup)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
	}

	//fetch all insolvency from backup if exist
	if len(insolvencyDaosFromBackup) > 0 {
		insolvencyFromBackup := transform(insolvencyDaosFromBackup)
		if len(insolvencyFromBackup) > 0 {
			// drop insolvency
			err = collection.Drop(context.Background())
			if err != nil {
				return nil, fmt.Errorf("failed to drop appointment collection")
			}
		}

		// transfer backup to insolvency
		_, err = collection.InsertMany(context.Background(), insolvencyFromBackup)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}

	}

	var insolvencyMigrationDaos []InsolvencyMigrationDao

	// Retrieve insolvency case from Mongo
	storedInsolvency, _ := collection.Find(context.Background(), bson.M{})
	err = storedInsolvency.All(context.Background(), &insolvencyMigrationDaos)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	// drop back up and recreate to prevent error
	err = insolvencyBackUpCollectionName.Drop(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to drop insolvency backup collection")
	}

	//backup insolvency incase of any issue if not exist
	insolvencyDaos := transform(insolvencyMigrationDaos)
	_, err = insolvencyBackUpCollectionName.InsertMany(context.Background(), insolvencyDaos)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	fmt.Println(insolvencyMigrationDaos)
	for _, insolvency := range insolvencyMigrationDaos {
		practitionersMapResource := make(map[string]string)
		practitioners := insolvency.Data.Practitioners
		var practitionerMigrationDaos []interface{}
		if len(practitioners) > 0 {
			for _, practitioner := range practitioners {
				practitionerModel := models.PractitionerResourceDao{}
				practitionerModel.Data.IPCode = practitioner.IPCode
				practitionerModel.Data.PractitionerId = practitioner.Id
				practitionerModel.Data.FirstName = practitioner.FirstName
				practitionerModel.Data.LastName = practitioner.LastName
				practitionerModel.Data.TelephoneNumber = practitioner.TelephoneNumber
				practitionerModel.Data.Email = practitioner.Email
				practitionerModel.Data.Address = models.AddressResourceDao{
					AddressLine1: practitioner.Address.AddressLine1,
					AddressLine2: practitioner.Address.AddressLine2,
					Premises:     practitioner.Address.Premises,
					Country:      practitioner.Address.Country,
					Locality:     practitioner.Address.Locality,
					Region:       practitioner.Address.Region,
					PostalCode:   practitioner.Address.PostalCode,
					POBox:        practitioner.Address.POBox,
				}
				practitionerModel.Data.Role = practitioner.Role
				practitionerModel.Data.Etag = insolvency.Etag
				practitionerModel.Data.Kind = insolvency.Kind
				practitionerModel.Data.Links = models.PractitionerResourceLinksDao{
					Self: insolvency.Links.Self,
				}
				if practitioner.Appointment != nil {
					practitionerMapAppointmentResource := make(map[string]string)
					practitionerMapAppointmentResource[practitioner.Id] = practitioner.Appointment.Links.Self
					appointmentLink, _ := utils.ConvertMapToString(practitionerMapAppointmentResource)
					practitionerModel.Data.Links.Appointment = appointmentLink
					appointment := models.AppointmentResourceDao{
						PractitionerId: practitioner.Id,
					}
					appointment.Data.AppointedOn = practitioner.Appointment.AppointedOn
					appointment.Data.MadeBy = practitioner.Appointment.MadeBy
					appointment.Data.Links = models.AppointmentResourceLinksDao{
						Self: practitioner.Appointment.Links.Self,
					}

					if len(practitioner.Appointment.Etag) == 0 {
						appointment.Data.Etag = generateEtag()
					} else {
						appointment.Data.Etag = practitioner.Appointment.Etag
					}

					if len(practitioner.Appointment.Kind) == 0 {
						appointment.Data.Kind = "insolvency#appointment"
					} else {
						appointment.Data.Kind = practitioner.Appointment.Kind
					}

					//save appointment
					_, err := appointmentCollectionName.InsertOne(context.Background(), appointment)
					if err != nil {
						fmt.Println("problem inserting appointment===>" + err.Error())
					}
				}
				//practitioners links
				practitionersMapResource[practitioner.Id] = fmt.Sprintf(constants.TransactionsPath + insolvency.TransactionID + constants.PractitionersPath + string(practitioner.Id))

				practitionerMigrationDaos = append(practitionerMigrationDaos, practitionerModel)
			}

			//TODO HERE
			// attachments
			// state-of-affairs
			//resolutions

			// save practitioners
			_, err := practitionerCollectionName.InsertMany(context.Background(), practitionerMigrationDaos)
			if err != nil {
				fmt.Println("problem inserting practitioner===>" + err.Error())
			}

			filter := bson.M{"_id": insolvency.ID}

			//remove practitioners
			updateDocument := bson.M{"$unset": bson.M{"data.practitioners": "", "etag": "", "kind": ""}}
			_, err = collection.UpdateOne(context.Background(), filter, updateDocument)
			if err != nil {
				fmt.Println("problem unset insolvency practitioner===>" + err.Error())
			}

			// add practitioner link to insolvency
			practitionersString, _ := utils.ConvertMapToString(practitionersMapResource)
			update := bson.M{"$set": bson.M{"data.practitioners": practitionersString, "data.etag": insolvency.Etag, "data.kind": insolvency.Kind}}
			_, err = collection.UpdateOne(context.Background(), filter, update)
			if err != nil {
				fmt.Println("problem update insolvency practitioner link===>" + err.Error())
			}
		}

	}
	// drop insolvency_back to refresh --run this only when you are sure all migration done successfully
	// err = insolvencyBackUpCollectionName.Drop(context.Background())
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to drop insolvency backup collection")
	// }

	return &insolvencyMigrationDaos, nil
}

func main() {
	// configure the address of the database you intended to run the migration here.
	cfg := &config.Config{
		BindAddr:        ":10092",
		MongoDBURL:      "mongodb://127.0.0.1:27017",
		Database:        "transactions_insolvency",
		MongoCollection: "insolvency",
	}

	svc := NewDAOMigrationService(cfg)

	insolvencyMigrationDao, err := svc.Migrate()
	if err != nil {
		fmt.Println("error:", err)
	}
	if insolvencyMigrationDao != nil {
		fmt.Println(insolvencyMigrationDao)
	}
}
