<script>
    import { onMount } from 'svelte'

    import DownCarrotIcon from '../components/icons/DownCarrotIcon.svelte'
    import SolidButton from '../components/SolidButton.svelte'
    import HollowButton from '../components/HollowButton.svelte'
    import { user } from '../stores.js'

    export let notifications
    export let router

    let storyboardName = ''

    function createStoryboard(e) {
        e.preventDefault()
        const data = {
            storyboardName,
        }

        fetch('/api/storyboard', {
            method: 'POST',
            credentials: 'same-origin',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        })
            .then(function(response) {
                return response.json()
            })
            .then(function(storyboard) {
                router.route(`/storyboard/${storyboard.id}`)
            })
            .catch(function(error) {
                notifications.danger('Error encountered creating storyboard')
            })
    }

    onMount(() => {
        if (!$user.id) {
            router.route('/register')
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
