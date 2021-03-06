<script>
    import PageLayout from '../components/PageLayout.svelte'
    import SolidButton from '../components/SolidButton.svelte'
    import { appRoutes } from '../config'
    import { user } from '../stores.js'

    export let xfetch
    export let router
    export let eventTag
    export let notifications
    export let resetId

    let userPassword1 = ''
    let userPassword2 = ''

    function resetUserPassword(e) {
        e.preventDefault()
        const body = {
            resetId,
            userPassword1,
            userPassword2,
        }
        const validPasswords = validatePasswords(userPassword1, userPassword2)

        let noFormErrors = true

        if (!validPasswords.valid) {
            noFormErrors = false
            notifications.danger(validPasswords.error, 1500)
        }

        if (noFormErrors) {
            xfetch('/api/auth/reset-password', { body })
                .then(function() {
                    eventTag('reset_password', 'engagement', 'success', () => {
                        router.route(appRoutes.login, true)
                    })
                })
                .catch(function(error) {
                    notifications.danger(
                        'Error encountered attempting to reset password',
                    )
                    eventTag('reset_password', 'engagement', 'failure')
                })
        }
    }

    $: resetDisabled = userPassword1 === '' || userPassword2 === ''
</script>

<svelte:head>
    <title>Reset Password | Exothermic</title>
</svelte:head>

<PageLayout>
    <div class="flex justify-center">
        <div class="w-full md:w-1/2 lg:w-1/3">
            <form
                on:submit="{resetUserPassword}"
                class="bg-white shadow-lg rounded p-6 mb-4"
                name="resetUserPassword">
                <div
                    class="font-bold text-xl md:text-2xl mb-2 md:mb-6
                    md:leading-tight text-center">
                    Reset Password
                </div>

                <div class="mb-4">
                    <label
                        class="block text-gray-700 text-sm font-bold mb-2"
                        for="yourPassword1">
                        Password
                    </label>
                    <input
                        bind:value="{userPassword1}"
                        placeholder="Enter a password"
                        class="bg-gray-200 border-gray-200 border-2
                        appearance-none rounded w-full py-2 px-3 text-gray-700
                        leading-tight focus:outline-none focus:bg-white
                        focus:border-orange-500"
                        id="yourPassword1"
                        name="yourPassword1"
                        type="password"
                        required />
                </div>

                <div class="mb-4">
                    <label
                        class="block text-gray-700 text-sm font-bold mb-2"
                        for="yourPassword2">
                        Confirm Password
                    </label>
                    <input
                        bind:value="{userPassword2}"
                        placeholder="Confirm your password"
                        class="bg-gray-200 border-gray-200 border-2
                        appearance-none rounded w-full py-2 px-3 text-gray-700
                        leading-tight focus:outline-none focus:bg-white
                        focus:border-orange-500"
                        id="yourPassword2"
                        name="yourPassword2"
                        type="password"
                        required />
                </div>

                <div class="text-right">
                    <SolidButton type="submit" disabled="{resetDisabled}">
                        Reset
                    </SolidButton>
                </div>
            </form>
        </div>
    </div>
</PageLayout>
