import type { Character } from "../api/generated";

export type ComparedUser = {
  id: string;
  discord_avatar?: string;
  discord_username: string;
};

export type UserWithCharacters = ComparedUser & {
  characters: Character[];
};

/** Character id -> ids of the users compared here who own it. */
export type Ownership = ReadonlyMap<string, readonly string[]>;

/** Users by id, so a card can resolve an owner without scanning. */
export type UserIndex = ReadonlyMap<string, ComparedUser>;

/** Character ids of a media, for membership checks in the grid filter. */
export type IdSet = ReadonlySet<number>;

export const buildUserIndex = (users: ComparedUser[]): UserIndex =>
  new Map(users.map((user) => [user.id, user]));

export const buildOwnership = (users: UserWithCharacters[]): Ownership => {
  const ownersByCharacter = new Map<string, Set<string>>();

  for (const user of users) {
    for (const char of user.characters) {
      const charId = char.id.toString();
      const owners = ownersByCharacter.get(charId);
      if (owners) owners.add(user.id);
      else ownersByCharacter.set(charId, new Set([user.id]));
    }
  }

  // Flattened once, here, rather than per card render.
  return new Map(Array.from(ownersByCharacter, ([charId, owners]) => [charId, Array.from(owners)]));
};

export const toIdSet = (characters: ReadonlyArray<{ id: number }> | undefined): IdSet =>
  new Set((characters ?? []).map((char) => char.id));
