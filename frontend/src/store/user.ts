import { createSignal } from "solid-js";
import type { Character, User, Collection } from "../api/generated";
import { getUser, getUserCollection } from "../api/generated";

interface UserState {
	user: User | null;
	characters: Character[] | null;
	loading: boolean;
	error: string | null;
}

export const useUserStore = () => {
	const [state, setState] = createSignal<UserState>({
		user: null,
		characters: null,
		loading: false,
		error: null,
	});

	const fetchProfile = async (userId: string) => {
		setState({ ...state(), loading: true, error: null });
		try {
			// Fetch user profile and collection in parallel
			const [userResult, collectionResult] = await Promise.all([
				getUser(userId),
				getUserCollection(userId),
			]);

			setState({
				user: userResult,
				characters: collectionResult.characters,
				loading: false,
				error: null,
			});
		} catch (error) {
			setState({
				user: null,
				characters: null,
				loading: false,
				error: error instanceof Error ? error.message : "Unknown error",
			});
			throw error;
		}
	};

	const fetchCharacters = async (userId: string) => {
		try {
			const result = await getUserCollection(userId);
			setState({
				...state(),
				characters: result.characters,
			});
			return result.characters;
		} catch (error) {
			setState({
				...state(),
				error: error instanceof Error ? error.message : "Unknown error",
			});
			throw error;
		}
	};

	const updateUser = (user: User) => {
		setState({
			user,
			characters: state().characters,
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
