<script>
    import Modal from './Modal.svelte'
    import SolidButton from './SolidButton.svelte'

    export let toggleCreate = () => {}
    export let handleCreate = () => {}

    export let departmentName = ''

    function onSubmit(e) {
        e.preventDefault()

        handleCreate(departmentName)
    }

    $: createDisabled = departmentName === ''
</script>

<Modal closeModal={toggleCreate}>
    <form on:submit="{onSubmit}" name="createDepartment">
        <div class="mb-4">
            <label
                class="block text-gray-700 text-sm font-bold mb-2"
                for="departmentName">
                Department Name
            </label>
            <input
                bind:value="{departmentName}"
                placeholder="Enter an department name"
                class="bg-gray-200 border-gray-200 border-2
                appearance-none rounded w-full py-2 px-3 text-gray-700
                leading-tight focus:outline-none focus:bg-white
                focus:border-purple-500"
                id="departmentName"
                name="departmentName"
                required />
        </div>

        <div>
            <div class="text-right">
                <SolidButton type="submit" disabled="{createDisabled}">
                    Create Department
                </SolidButton>
            </div>
        </div>
    </form>
</Modal>
