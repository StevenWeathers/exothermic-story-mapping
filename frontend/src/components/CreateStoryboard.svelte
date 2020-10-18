<script>
    import { onMount } from 'svelte'

    import SolidButton from '../components/SolidButton.svelte'
    import { user } from '../stores.js'
    import { appRoutes } from '../config'

    export let xfetch
    export let notifications
    export let eventTag
    export let router

    let storyboardName = ''

    function createStoryboard(e) {
        e.preventDefault()
        const body = {
            storyboardName,
        }

        xfetch('/api/storyboard', { body })
            .then(res => res.json())
            .then(function(storyboard) {
                eventTag('create_storyboard', 'engagement', 'success', () => {
                    router.route(`${appRoutes.storyboard}/${storyboard.id}`)
                })
            })
            .catch(function(error) {
                notifications.danger('Error encountered creating storyboard')
                eventTag('create_storyboard', 'engagement', 'failure')
            })
    }

    onMount(() => {
        if (!$user.id) {
            router.route(appRoutes.register)
        }
    })
</script>

<form on:submit="{createStoryboard}" name="createStoryboard">
    <div class="mb-4">
        <label
            class="block text-gray-700 text-sm font-bold mb-2"
            for="storyboardName">
            Storyboard Name
        </label>
        <div class="control">
            <input
                name="storyboardName"
                bind:value="{storyboardName}"
                placeholder="Enter a storyboard name"
                class="bg-gray-200 border-gray-200 border-2 appearance-none
                rounded w-full py-2 px-3 text-gray-700 leading-tight
                focus:outline-none focus:bg-white focus:border-orange-500"
                id="storyboardName"
                required />
        </div>
    </div>

    <div class="text-right">
        <SolidButton type="submit">Create Storyboard</SolidButton>
    </div>
</form>
