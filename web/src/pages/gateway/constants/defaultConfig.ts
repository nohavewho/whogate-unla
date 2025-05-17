import { getRandomLetters } from "../../../utils/utils";
import { GatewayConfig } from '../types';

// Default configuration object for new or empty configurations
export const defaultConfig: GatewayConfig = {
  name: "",
  tenant: "default",
  routers: [{
    server: "",
    prefix: "/" + getRandomLetters(4)
  }],
  servers: [{
    name: "",
    description: "",
    allowedTools: []
  }],
  tools: []
};

// Default MCP configuration object
export const defaultMCPConfig: GatewayConfig = {
  name: "",
  tenant: "default",
  routers: [{
    server: "",
    prefix: "/mcp"
  }],
  mcpServers: [{
    type: "stdio",
    name: "",
    command: "",
    args: [],
    env: {}
  }]
}; 