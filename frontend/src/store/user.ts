import { createSignal } from "solid-js";
import type { Character, Profile } from "../api/generated";
import { getUserV1 } from "../api/generated";

interface UserState {
	profile: Profile | null;
	characters: Character[] | null;
	loading: boolean;
	error: string | null;
}

export const useUserStore = () => {
	const [state, setState] = createSignal<UserState>({
		profile: null,
		characters: null,
		loading: false,
		error: null,
	});

	const fetchProfile = async (userId: string) => {
		setState({ ...state(), loading: true, error: null });
		try {
			const result = await getUserV1(userId);
			setState({
				profile: result,
				characters: result.waifus,
				loading: false,
				error: null,
			});
		} catch (error) {
			setState({
				profile: null,
				characters: null,
				loading: false,
				error: error instanceof Error ? error.message : "Unknown error",
			});
			throw error;
		}
	};

	const fetchCharacters = async (userId: string) => {
		try {
			const result = await getUserV1(userId);
			setState({
				...state(),
				characters: result.waifus,
			});
			return result.waifus;
		} catch (error) {
			setState({
				...state(),
				error: error instanceof Error ? error.message : "Unknown error",
			});
			throw error;
		}
	};

	const updateUser = (profile: Profile) => {
		setState({
			profile,
			characters: profile.waifus,
			loading: state().loading,
			error: state().error,
		});
	};

	const setLoading = (loading: boolean) => {
		setState({
			...state(),
			loading,
		});
	};

	const setError = (error: string | null) => {
		setState({
			...state(),
			error,
		});
	};

	return {
		get state() {
			return state();
		},
		fetchProfile,
		fetchCharacters,
		updateUser,
		setLoading,
		setError,
	};
};
