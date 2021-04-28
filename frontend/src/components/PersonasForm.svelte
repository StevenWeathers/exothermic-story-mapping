<script>
    import SolidButton from './SolidButton.svelte'
    import CloseIcon from './icons/CloseIcon.svelte'

    export let toggleEditPersona = () => () => {}
    export let handlePersonaAdd = () => {}
    export let handlePersonaRevision = () => {}

    export let persona = {
        id: '',
        name: '',
        role: '',
        description: ''
    }

    function handleSubmit(event) {
        event.preventDefault()

        if (persona.id === '') {
            handlePersonaAdd({
                name: persona.name,
                role: persona.role,
                description: persona.description
            })
        } else {
            handlePersonaRevision(persona)
        }
        toggleEditPersona()
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
                    on:click="{toggleEditPersona}"
                    class="text-gray-800">
                    <CloseIcon />
                </button>
            </div>

            <form on:submit="{handleSubmit}" name="addPersona">
                <div class="mb-4">
                    <label class="block text-sm font-bold mb-2" for="personaName">
                        Persona Name
                    </label>
                    <input
                        class="bg-gray-200 border-gray-200 border-2
                        appearance-none rounded w-full py-2 px-3 text-gray-700
                        leading-tight focus:outline-none focus:bg-white
                        focus:border-purple-500"
                        id="personaName"
                        type="text"
                        bind:value="{persona.name}"
                        placeholder="Enter a persona name e.g. Ricky Bobby"
                        name="personaName" />
                </div>
                <div class="mb-4">
                    <label class="block text-sm font-bold mb-2" for="personaRole">
                        Persona Role
                    </label>
                    <input
                        class="bg-gray-200 border-gray-200 border-2
                        appearance-none rounded w-full py-2 px-3 text-gray-700
                        leading-tight focus:outline-none focus:bg-white
                        focus:border-purple-500"
                        id="personaRole"
                        type="text"
                        bind:value="{persona.role}"
                        placeholder="Enter a persona role e.g. Author, Developer, Admin"
                        name="personaRole" />
                </div>
                <div class="mb-4">
                    <label class="block text-sm font-bold mb-2" for="personaDescription">
                        Persona Description
                    </label>
                    <textarea
                        class="bg-gray-200 border-gray-200 border-2
                        appearance-none rounded w-full py-2 px-3 text-gray-700
                        leading-tight focus:outline-none focus:bg-white
                        focus:border-purple-500"
                        id="personaDescription"
                        bind:value="{persona.description}"
                        placeholder="Enter a persona description"
                        name="personaDescription" />
                </div>
                <div class="text-right">
                    <div>
                        <SolidButton type="submit">Save</SolidButton>
                    </div>
                </div>
            </form>
        </div>
    </div>
</div>
