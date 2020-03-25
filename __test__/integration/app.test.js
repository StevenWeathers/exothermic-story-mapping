const appUrl = process.env.APP_URL || 'http://localhost:8080/'

describe('Exothermic App', () => {
    beforeAll(async () => {
        await page.goto(appUrl) // @TODO - make this configurable for the endpoint
    })

    describe('when on the Landing page', () => {
        it('should match a link with a "Create a Storyboard" text inside', async () => {
            await expect(page).toMatchElement('a', {
                text: 'Create a Storyboard',
            })
        })
    })

    describe('when clicking the Create a Storyboard button as a new visitor', () => {
        it('should redirect to Enlist page', async () => {
            await expect(page).toClick('a', { text: 'Create a Storyboard' })
            await expect(page).toMatchElement('input[name="yourName1"]')
        })

        it('should match an input with a "yourName" name then fill it with text', async () => {
            await expect(page).toFill('input[name="yourName1"]', 'Thor')
        })

        it('should then submit the form and be redirected to Storyboards page', async () => {
            await expect(page).toClick(
                'form[name="registerGuest"] button[type="submit"]',
            )
            await page.waitFor(1000) // give the page a little time to redirect and render
            await expect(page).toMatchElement('h1', { text: 'My Storyboards' })
        })
    })

    describe('when on the My Storyboards page', () => {
        it('should match an input with a "storyboardName" name then fill it with text', async () => {
            await expect(page).toFill('input[name="storyboardName"]', 'Asgard')
        })

        it("should then submit the form and be redirected to that new Storyboard's page", async () => {
            await expect(page).toClick(
                'form[name="createStoryboard"] button[type="submit"]',
            )
            await page.waitFor(1000) // give the page a little time to redirect and render
            await expect(page).toMatchElement('h1', { text: 'Asgard' })
        })
    })

    describe('when on a Storyboard page as owner', () => {
        it('should have Delete Storyboard Button', async () => {
            await expect(page).toMatchElement('button', {
                text: 'Delete Storyboard',
            })
        })

        describe('when clicking the Delete Storyboard button', () => {
            it('should redirect to Landing page', async () => {
                await expect(page).toClick('button', {
                    text: 'Delete Storyboard',
                })
                await page.waitFor(1000) // wait for page to redirect
                await expect(page).toMatchElement('a', {
                    text: 'Create a Storyboard',
                })
            })
        })
    })
})
