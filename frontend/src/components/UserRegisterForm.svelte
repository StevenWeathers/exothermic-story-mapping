<script>
    import SolidButton from '../components/SolidButton.svelte'
    import { validateName, validatePasswords } from '../validationUtils.js'
    import { _ } from '../i18n'

    export let notifications
    export let handleSubmit
    export let guestUsersName = ''

    let userName = guestUsersName
    let userEmail = ''
    let userPassword1 = ''
    let userPassword2 = ''

    function onSubmit(e) {
        e.preventDefault()

        const validName = validateName(userName)
        const validPasswords = validatePasswords(userPassword1, userPassword2)

        let noFormErrors = true

        if (!validName.valid) {
            noFormErrors = false
            notifications.danger(validName.error, 1500)
        }

        if (!validPasswords.valid) {
            noFormErrors = false
            notifications.danger(validPasswords.error, 1500)
        }

        if (noFormErrors) {
            handleSubmit(userName, userEmail, userPassword1, userPassword2)
        }
    }

    $: createDisabled =
        userName === '' ||
        userEmail === '' ||
        userPassword1 === '' ||
        userPassword2 === ''
</script>

<form on:submit="{onSubmit}" name="createAccount">

    <div class="mb-4">
        <label
            class="block text-gray-700 text-sm font-bold mb-2"
            for="yourName2">
            Name
        </label>
        <input
            bind:value="{userName}"
            placeholder="Enter your name"
            class="bg-gray-200 border-gray-200 border-2 appearance-none rounded
            w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none
            focus:bg-white focus:border-orange-500"
            id="yourName2"
            name="yourName2"
            required />
    </div>

    <div class="mb-4">
        <label
            class="block text-gray-700 text-sm font-bold mb-2"
            for="yourEmail">
            Email
        </label>
        <input
            bind:value="{userEmail}"
            placeholder="Enter your email"
            class="bg-gray-200 border-gray-200 border-2 appearance-none rounded
            w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none
            focus:bg-white focus:border-orange-500"
            id="yourEmail"
            name="yourEmail"
            type="email"
            required />
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
            class="bg-gray-200 border-gray-200 border-2 appearance-none rounded
            w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none
            focus:bg-white focus:border-orange-500"
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
            class="bg-gray-200 border-gray-200 border-2 appearance-none rounded
            w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none
            focus:bg-white focus:border-orange-500"
            id="yourPassword2"
            name="yourPassword2"
            type="password"
            required />
    </div>

    <div>
        <div class="text-right">
            <SolidButton type="submit" disabled="{createDisabled}">
                Create
            </SolidButton>
        </div>
    </div>
</form>
