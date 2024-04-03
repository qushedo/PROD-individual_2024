package travel

import (
	"backend-qushedo/functions"
	tb "gopkg.in/telebot.v3"
)

func SetupTravels(b *tb.Bot) {
	//Terrible functions structure, I won't redo it, well, that's fine
	travelGroup := b.Group()
	travelGroup.Handle(&btnBackTravels, functions.MainMenuWithDelete)
	travelGroup.Handle(&btnBackTravel, MyTravels)
	travelGroup.Handle(&btnBackEditTravel, OpenTravelMenu)
	travelGroup.Handle(&btnBackLocations, OpenTravelMenu)
	travelGroup.Handle(&btnBackLocation, LocationsMenu)
	travelGroup.Handle(&btnBackMembersList, OpenTravelMenu)
	travelGroup.Handle(&btnBackMemberMenu, functions.Back)

	travelGroup.Handle(&functions.BtnMyTravels, MyTravels)

	travelGroup.Handle(&btnCreateNewTravel, NewTravel)

	travelGroup.Handle(&btnEditTravel, EditTravel)

	travelGroup.Handle(&btnEditTravelLocations, LocationsMenu)
	travelGroup.Handle(&btnLocationDeleteNo, OpenTravelMenu)
	travelGroup.Handle(&btnEditTravelMembers, MembersMenu)

	travelGroup.Use(IsTravelOwnerMiddleware)
	travelGroup.Handle(&btnEditTravelName, EditTravelName)
	travelGroup.Handle(&btnEditTravelDesc, EditTravelDesc)
	travelGroup.Handle(&btnDeleteTravel, DeleteTravel)
	travelGroup.Handle(&btnTravelDeleteYes, YesDeleteTravel)
	travelGroup.Handle(&btnTravelDeleteNo, MyTravels)

	travelGroup.Handle(&btnAddNewLocation, AddLocation)
	travelGroup.Handle(&btnCorrectLocation, LocationCorrect)
	travelGroup.Handle(&btnIncorrectLocation, LocationIncorrect)
	travelGroup.Handle(&btnDeleteLocation, DeleteLocation)
	travelGroup.Handle(&btnLocationDeleteYes, YesDeleteLocation)

	travelGroup.Handle(&btnMemberKick, KickFromTravel)
	travelGroup.Handle(&btnGenerateInvite, GenerateInviteLink)

}
