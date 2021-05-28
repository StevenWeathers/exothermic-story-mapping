<script>
    import Modal from './Modal.svelte'
    import SolidButton from './SolidButton.svelte'

    export let toggleCreate = () => {}
    export let handleCreate = () => {}

    export let organizationName = ''

    function onSubmit(e) {
        e.preventDefault()

        handleCreate(organizationName)
    }

    $: createDisabled = organizationName === ''
</script>

<Modal closeModal="{toggleCreate}">
    <form on:submit="{onSubmit}" name="createOrganization">
        <div class="mb-4">
            <label
                class="block text-gray-700 text-sm font-bold mb-2"
                for="organizationName">
                Organization Name
            </label>
            <input
                bind:value="{organizationName}"
                placeholder="Enter an organization name"
                class="bg-gray-200 border-gray-200 border-2 appearance-none
                rounded w-full py-2 px-3 text-gray-700 leading-tight
                focus:outline-none focus:bg-white focus:border-purple-500"
                id="organizationName"
                name="organizationName"
                required />
        </div>

        <div>
            <div class="text-right">
                <SolidButton type="submit" disabled="{createDisabled}">
                    Create Organization
                </SolidButton>
            </div>
        </div>
    </form>
</Modal>
