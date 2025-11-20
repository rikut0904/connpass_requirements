export type GuildPermission = {
  id: number;
  guildId: string;
  guildName: string;
  permissions: number;
  iconUrl?: string;
  canManage: boolean;
  canManageRole: boolean;
};

export type Rule = {
  id: number;
  guildId: string;
  channelId: string;
  channelName: string;
  name: string;
  description: string;
  location: string;
  capacityThreshold: number;
  keywords: string[];
  notifyTypes: string[];
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

export type ImportantLog = {
  id: number;
  level: string;
  eventType: string;
  message: string;
  metadata?: string;
  createdAt: string;
};

export type SchedulerStatus = {
  id: number;
  lastRunAt?: string;
  lastError?: string;
  updatedAt?: string;
};
