<script>
    import SolidButton from './SolidButton.svelte'
    import CloseIcon from './icons/CloseIcon.svelte'
    import HollowButton from './HollowButton.svelte'

    export let toggleColumnEdit = () => {}
    export let handleColumnRevision = () => {}
    export let deleteColumn = () => () => {}

    export let column = {
        id: '',
        name: '',
    }

    function handleSubmit(event) {
        event.preventDefault()

        handleColumnRevision(column)
        toggleColumnEdit()
    }
</script>

<div class="fixed inset-0 flex items-center z-40">
    <div class="fixed inset-0 bg-gray-900 opacity-75"></div>

    <div
        class="relative mx-4 md:mx-auto w-full md:w-2/3 lg:w-3/5 xl:w-1/3 z-50
        m-8">
        <div class="shadow-xl bg-white rounded-lg p-4 xl:p-6">
            <div class="flex justify-end mb-2">
                <button
                    aria-label="close"
                    on:click="{toggleColumnEdit}"
                    class="text-gray-800">
                    <CloseIcon />
                </button>
            </div>

            <form on:submit="{handleSubmit}" name="addColumn">
                <div class="mb-4">
                    <label
                        class="block text-sm font-bold mb-2"
                        for="columnName">
                        Column Name
                    </label>
                    <input
                        class="bg-gray-200 border-gray-200 border-2
                        appearance-none rounded w-full py-2 px-3 text-gray-700
                        leading-tight focus:outline-none focus:bg-white
                        focus:border-purple-500"
                        id="columnName"
                        type="text"
                        bind:value="{column.name}"
                        placeholder="Enter a column name"
                        name="columnName" />
                </div>
                <div class="flex">
                    <div class="md:w-1/2 text-left">
                        <HollowButton
                            color="red"
                            onClick="{deleteColumn(column.id)}">
                            Delete Column
                        </HollowButton>
                    </div>
                    <div class="md:w-1/2 text-right">
                        <SolidButton type="submit">Save</SolidButton>
                    </div>
                </div>
            </form>
        </div>
    </div>
</div>
