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

	"github.com/companieshouse/insolvency-api/config"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// This can be run using the command go run ./dao/migration/migrate.go depending on your file structure.
//
// The following environment variables are required (example values, replace with correct for environment):
// MONGODB_URL="mongodb://127.0.0.1:27017"
// INSOLVENCY_MONGODB_DATABASE="transactions_insolvency"
// INSOLVENCY_MONGODB_COLLECTION="insolvency"
//
// A new backup is taken on each run to an insolvency_backup_<DATETIME> collection
// If errors are encountered at any point it would stop without completing.
// Please look at whatever error message it outputs and rectify.
// Before re-attempting, carefully review the output & ensure that the data is in the expected starting state:
// - insolvency colection contains old-format data only (may need to restore from insolvency_backup_<DATETIME>)
// - practitioners & appointments collections are not present
//
// Please note some models are recreated in order to be able to complete the migration without distorting the documents.
// Also note this script only includes the new practitioners and appointments collections,
// a further similar script will need to be created for onward work on other new collections.

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
		fmt.Println(fmt.Errorf("error connecting to mongodb: %s. Exiting", err))
		os.Exit(1)
	}

	// check we can connect to the mongodb instance. failure here should result in a crash.
	pingContext, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
	defer cancel()
	err = client.Ping(pingContext, nil)
	if err != nil {
		fmt.Println(errors.New("ping to mongodb timed out. please check the connection to mongodb and that it is running"))
		os.Exit(1)
	}

	fmt.Println("connected to mongodb successfully")

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

// get count of documents in a collection

func getCollectionCount(collection *mongo.Collection) (int64, error) {
	count, err := collection.CountDocuments(context.Background(), bson.D{})
	if err != nil {
		return 0, fmt.Errorf("failed to get count of [%s] collection: [%s]", collection.Name(), err)
	}
	return count, nil
}

func (m *MongoMigrationService) Migrate() (*[]InsolvencyMigrationDao, error) {

	insolvencyCollection := m.db.Collection(m.CollectionName)
	practitionerCollection := m.db.Collection(dao.PractitionerCollectionName)
	appointmentCollection := m.db.Collection(dao.AppointmentCollectionName)
	t := time.Now()
	insolvencyBackUpCollectionName := "insolvency_backup_" + t.Format("20060102150405")
	insolvencyBackUpCollection := m.db.Collection(insolvencyBackUpCollectionName)

	// check for data in new collections and abort if already present
	count, err := getCollectionCount(practitionerCollection)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("[%s] already populated (count = [%v]), has migration script already been run?", dao.PractitionerCollectionName, count)
	}
	count, err = getCollectionCount(appointmentCollection)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("[%s] already populated (count = [%v]), has migration script already been run?", dao.AppointmentCollectionName, count)
	}

	// check for new-format data in insolvency collection and abort if already present
	count, err = insolvencyCollection.CountDocuments(context.Background(), bson.D{{Key: "data.etag", Value: bson.M{"$exists": 1}}})
	if err != nil {
		return nil, fmt.Errorf("failed to get count of [%s] collection where 'data.etag' exists: [%s]", insolvencyCollection.Name(), err)
	}
	if count > 0 {
		return nil, fmt.Errorf("[%s] already contains updated format 'data.etag' (count = [%v]), has migration script already been run?", insolvencyCollection.Name(), count)
	}

	var insolvencyMigrationDaos []InsolvencyMigrationDao

	// Retrieve insolvency cases from Mongo
	storedInsolvency, _ := insolvencyCollection.Find(context.Background(), bson.M{})
	err = storedInsolvency.All(context.Background(), &insolvencyMigrationDaos)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	fmt.Printf("existing insolvency collection retrieved ([%v] docs)\n", len(insolvencyMigrationDaos))

	//back up insolvency collection using aggregation with only an '$out' stage
	pipeline := []bson.D{{{Key: "$out", Value: insolvencyBackUpCollectionName}}}
	_, err = insolvencyCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	// check that backup collection is the expected size before proceeding
	count, err = getCollectionCount(insolvencyBackUpCollection)
	if err != nil {
		return nil, err
	}
	if count != int64(len(insolvencyMigrationDaos)) {
		return nil, fmt.Errorf("insolvency ([%v]) & [%s] ([%v]) counts don't match ", len(insolvencyMigrationDaos), insolvencyBackUpCollectionName, count)
	}
	fmt.Printf("existing insolvency collection backed up to [%s] ([%v] docs)\n\n", insolvencyBackUpCollectionName, count)

	for _, insolvency := range insolvencyMigrationDaos {
		fmt.Printf("starting to process insolvency resource data for transactionID [%s]\n", insolvency.TransactionID)
		practitionersMapResource := make(map[string]string)
		practitioners := insolvency.Data.Practitioners
		var practitionerMigrationDaos []interface{}
		if len(practitioners) > 0 {
			for _, practitioner := range practitioners {
				practitionerModel := models.PractitionerResourceDao{
					TransactionID: insolvency.TransactionID,
				}
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
				practitionerModel.Data.Kind = "insolvency#practitioner"
				practitionerModel.Data.Links = models.PractitionerResourceLinksDao{
					Self: practitioner.Links.Self,
				}
				if practitioner.Appointment != nil {
					practitionerModel.Data.Links.Appointment = practitioner.Appointment.Links.Self
					appointment := models.AppointmentResourceDao{
						PractitionerId: practitioner.Id,
						TransactionID:  insolvency.TransactionID,
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
					_, err := appointmentCollection.InsertOne(context.Background(), appointment)
					if err != nil {
						return nil, fmt.Errorf("problem inserting appointment===>" + err.Error())
					}
					fmt.Printf("--- appointment saved for practitionerID [%s]\n", practitioner.Id)
				}
				//practitioners links
				practitionersMapResource[practitioner.Id] = fmt.Sprintf(constants.TransactionsPath + insolvency.TransactionID + constants.PractitionersPath + string(practitioner.Id))
				fmt.Printf("-- practitioner resource data prepared for practitionerID [%s]\n", practitioner.Id)

				practitionerMigrationDaos = append(practitionerMigrationDaos, practitionerModel)
			}

			// save practitioners
			practitionerResult, err := practitionerCollection.InsertMany(context.Background(), practitionerMigrationDaos)
			if err != nil {
				return nil, fmt.Errorf("problem inserting practitioner===>" + err.Error())
			}
			fmt.Printf("- practitioners saved for transactionID [%s] ([%v] docs)\n", insolvency.TransactionID, len(practitionerResult.InsertedIDs))

		} else {
			fmt.Printf("- insolvency resource for transactionID [%s] has no practitioners\n", insolvency.TransactionID)
		}
		filter := bson.M{"_id": insolvency.ID}
		unsetUpdateDocument := bson.M{"$unset": bson.M{"etag": "", "kind": "", "links": ""}}
		setUpdateDocument := bson.M{"$set": bson.M{"data.etag": insolvency.Etag, "data.kind": insolvency.Kind, "data.links": insolvency.Links}}
		if len(practitioners) > 0 {
			unsetUpdateDocument = bson.M{"$unset": bson.M{"data.practitioners": "", "etag": "", "kind": "", "links": ""}}
			setUpdateDocument = bson.M{"$set": bson.M{"data.practitioners": practitionersMapResource, "data.etag": insolvency.Etag, "data.kind": insolvency.Kind, "data.links": insolvency.Links}}
		}
		//unset operation (practitioners if present, etag, kind & links for all docs)
		_, err = insolvencyCollection.UpdateOne(context.Background(), filter, unsetUpdateDocument)
		if err != nil {
			return nil, fmt.Errorf("problem unsetting data for insolvency resource ===>" + err.Error())
		}

		// set operation (practitioners if present, etag, kind & links for all docs)
		_, err = insolvencyCollection.UpdateOne(context.Background(), filter, setUpdateDocument)
		if err != nil {
			return nil, fmt.Errorf("problem updating data for insolvency resource ===>" + err.Error())
		}
		fmt.Printf("insolvency resource data updated for transactionID [%s]\n\n", insolvency.TransactionID)

	}

	// for each created or updated collection, print out final count
	collections := []*mongo.Collection{insolvencyBackUpCollection, appointmentCollection, practitionerCollection, insolvencyCollection}
	for _, coll := range collections {
		count, err = getCollectionCount(coll)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			fmt.Printf("collection [%s] contains [%v] docs\n", coll.Name(), count)
		}
		// check that the insolvency collection's final count matches count of migrated docs
		if coll.Name() == insolvencyCollection.Name() {
			if count > int64(len(insolvencyMigrationDaos)) {
				return nil, fmt.Errorf("final validation check failed - collection [%s] count [%v] > migration update count of [%v] - has new data come in?", coll.Name(), count, len(insolvencyMigrationDaos))
			} else if count < int64(len(insolvencyMigrationDaos)) {
				return nil, fmt.Errorf("final validation check failed - collection [%s] count [%v] < migration update count of [%v] - has data been lost?", coll.Name(), count, len(insolvencyMigrationDaos))
			}
		}
	}

	return &insolvencyMigrationDaos, nil
}

func main() {

	// Get environment config
	cfg, err := config.Get()
	if err != nil {
		fmt.Println(fmt.Errorf("error configuring service: %s. Exiting", err), nil)
		return
	}
	if cfg.MongoDBURL == "" || cfg.Database == "" || cfg.MongoCollection == "" {
		fmt.Println(fmt.Errorf("config required: MONGODB_URL, INSOLVENCY_MONGODB_DATABASE, INSOLVENCY_MONGODB_COLLECTION. Exiting"))
		return
	}

	svc := NewDAOMigrationService(cfg)

	insolvencyMigrationDao, err := svc.Migrate()
	if err != nil {
		fmt.Println("error:", err)
		fmt.Println("Migration *** NOT ***  completed")
		return
	}

	if insolvencyMigrationDao != nil && len(*insolvencyMigrationDao) > 0 {
		fmt.Println("Migration completed")
		fmt.Println()
	} else {
		fmt.Println("Migration *** NOT ***  completed - no insolvency records processed")
	}
}
